package bot

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/AlibekAbdunasimov/chiptatop/internal/config"
	"github.com/AlibekAbdunasimov/chiptatop/internal/services/train"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api          *tgbotapi.BotAPI
	cfg          config.Config
	trainService *train.Service
	userStates   map[int64]*UserState
}

type UserState struct {
	CurrentStep string
	FromStation string
	ToStation   string
	SearchDate  time.Time
}

func New(cfg config.Config) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(cfg.TelegramBotToken)
	if err != nil {
		return nil, err
	}

	// Initialize train service with default language (Uzbek)
	trainService := train.NewService()

	// Try to use environment credentials first, otherwise initialize dynamically
	if cfg.RailwayXSRFToken != "" && cfg.RailwayCookies != "" {
		trainService.SetAuthCredentials(cfg.RailwayXSRFToken, cfg.RailwayCookies)
		log.Printf("Railway API authentication configured from environment")
	} else {
		log.Printf("No environment credentials - initializing dynamically...")
		// Initialize credentials dynamically
		if err := trainService.InitializeCredentials(context.Background()); err != nil {
			log.Printf("Warning: Failed to initialize credentials dynamically: %v", err)
			log.Printf("Train searches will fail until credentials are obtained")
		} else {
			log.Printf("Railway API authentication initialized dynamically")
		}
	}

	log.Printf("Bot @%s started in %s environment", api.Self.UserName, cfg.Environment)
	return &Bot{
		api:          api,
		cfg:          cfg,
		trainService: trainService,
		userStates:   make(map[int64]*UserState),
	}, nil
}

func (b *Bot) Run(ctx context.Context) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30

	updates := b.api.GetUpdatesChan(u)

	// Graceful shutdown handling
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-stop:
			log.Println("Shutting down bot...")
			return nil
		case update := <-updates:
			// Handle callback queries (inline keyboard buttons)
			if update.CallbackQuery != nil {
				b.handleCallbackQuery(update)
				continue
			}

			if update.Message == nil {
				continue
			}

			// Handle commands
			if update.Message.IsCommand() {
				b.handleCommand(update)
				continue
			}

			// Handle text messages (menu button clicks)
			if update.Message.Text != "" {
				b.handleTextMessage(update)
				continue
			}
		}
	}
}

// handleCallbackQuery handles all callback queries from inline keyboards
func (b *Bot) handleCallbackQuery(update tgbotapi.Update) {
	callback := update.CallbackQuery
	data := callback.Data

	// Answer the callback query to remove loading state
	b.api.Request(tgbotapi.NewCallback(callback.ID, ""))

	if strings.HasPrefix(data, "month_") || strings.HasPrefix(data, "date_") {
		b.handleCalendarCallback(update)
	} else if data == "main_menu" {
		// Handle main menu button from inline keyboard
		b.handleMainMenuButton(callback.Message.Chat.ID)
	} else if data == "header" || data == "empty" {
		// These are non-actionable buttons, just ignore them
		return
	}
}

func (b *Bot) handleCommand(update tgbotapi.Update) {
	switch update.Message.Command() {
	case "start":
		b.handleStartCommand(update)
	case "help":
		b.handleHelpCommand(update)
	case "stations":
		b.handleStationsCommand(update)
	case "search":
		b.handleSearchCommand(update)
	case "search_date":
		b.handleSearchDateCommand(update)
	default:
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Unknown command. Try /help to see available commands.")
		b.safeSend(msg)
	}
}

func (b *Bot) handleStartCommand(update tgbotapi.Update) {
	welcomeText := `🚂 *Welcome to ChiptaTop!*

I will help you find train tickets instantly. Use the menu buttons below:`

	// Create main menu keyboard
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔍 Search Trains"),
			tgbotapi.NewKeyboardButton("📅 Search by Date"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🚉 View Stations"),
			tgbotapi.NewKeyboardButton("🌍 Change Language"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("❓ Help"),
		),
	)
	keyboard.ResizeKeyboard = true
	keyboard.OneTimeKeyboard = false

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, welcomeText)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard
	b.safeSend(msg)
}

func (b *Bot) handleHelpCommand(update tgbotapi.Update) {
	helpText := `🚂 *ChiptaTop Train Bot Help*

🔍 *How to Use:*
• Use the menu buttons to navigate
• Search for trains between any stations
• View available dates and times
• Change language as needed

📋 *Available Options:*
• Search Trains - Find trains for today
• Search by Date - Find trains for specific date
• View Stations - See all available stations
• Change Language - Switch between Uzbek/Russian/English

💡 *Tips:*
• All major cities are supported
• Results show available seats and prices
• Automatic language detection`

	// Create help keyboard with back button
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔙 Back to Main Menu"),
		),
	)
	keyboard.ResizeKeyboard = true
	keyboard.OneTimeKeyboard = false

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, helpText)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard
	b.safeSend(msg)
}

func (b *Bot) handleTextMessage(update tgbotapi.Update) {
	text := update.Message.Text
	chatID := update.Message.Chat.ID

	// Get or create user state
	userState := b.getUserState(chatID)

	switch text {
	case "🔍 Search Trains":
		b.handleSearchTrainsButton(chatID)
	case "📅 Search by Date":
		b.handleSearchByDateButton(chatID)
	case "🚉 View Stations":
		b.handleViewStationsButton(chatID)
	case "🌍 Change Language":
		b.handleChangeLanguageButton(chatID)
	case "❓ Help":
		b.handleHelpButton(chatID)
	case "🔙 Back to Main Menu":
		b.handleMainMenuButton(chatID)
	case "🇺🇿 O'zbekcha":
		b.handleLanguageChange(chatID, "uz")
	case "🇷🇺 Русский":
		b.handleLanguageChange(chatID, "ru")
	case "🇺🇸 English":
		b.handleLanguageChange(chatID, "en")
	default:
		// Handle station selection based on current step
		if userState.CurrentStep != "" {
			b.handleStationSelection(chatID, text, userState)
			return
		}

		// Check if it's a search request (legacy support)
		if strings.Contains(text, " ") {
			parts := strings.Fields(text)
			if len(parts) == 2 {
				// Format: "from to" - search for today
				b.handleSearchRequest(chatID, parts[0], parts[1], time.Now())
				return
			} else if len(parts) == 3 {
				// Format: "from to date" - search for specific date
				date, err := time.Parse("2006-01-02", parts[2])
				if err == nil {
					b.handleSearchRequest(chatID, parts[0], parts[1], date)
					return
				}
			}
		}

		// Unknown text, show help
		msg := tgbotapi.NewMessage(chatID,
			"❓ I didn't understand that. Please use the menu buttons or send a search request in the format:\n\n"+
				"`from_station to_station`\n"+
				"or\n"+
				"`from_station to_station YYYY-MM-DD`")
		msg.ParseMode = "Markdown"
		b.safeSend(msg)
	}
}

func (b *Bot) getUserState(chatID int64) *UserState {
	if state, exists := b.userStates[chatID]; exists {
		return state
	}

	// Create new user state
	state := &UserState{
		CurrentStep: "",
		FromStation: "",
		ToStation:   "",
		SearchDate:  time.Time{},
	}
	b.userStates[chatID] = state
	return state
}

func (b *Bot) resetUserState(chatID int64) {
	b.userStates[chatID] = &UserState{
		CurrentStep: "",
		FromStation: "",
		ToStation:   "",
		SearchDate:  time.Time{},
	}
}

func (b *Bot) handleStationSelection(chatID int64, text string, userState *UserState) {
	switch userState.CurrentStep {
	case "select_from_station":
		// Extract station name from button text (no flag emoji anymore)
		stationName := strings.TrimSpace(text)

		// Debug logging
		log.Printf("DEBUG: Button text: '%s', Extracted station: '%s'", text, stationName)

		// Validate station name
		if stationName == "" {
			msg := tgbotapi.NewMessage(chatID, "❌ Invalid station selection. Please try again.")
			b.safeSend(msg)
			return
		}

		// Store the clean station name
		userState.FromStation = stationName
		userState.CurrentStep = "select_to_station"

		// Show destination station selection
		msg := tgbotapi.NewMessage(chatID,
			fmt.Sprintf("✅ Departure station: *%s*\n\nNow select your destination station:", stationName))
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Toshkent"),
				tgbotapi.NewKeyboardButton("Samarqand"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Buxoro"),
				tgbotapi.NewKeyboardButton("Andijon"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Qarshi"),
				tgbotapi.NewKeyboardButton("Termiz"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Nukus"),
				tgbotapi.NewKeyboardButton("Xiva"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Jizzax"),
				tgbotapi.NewKeyboardButton("Navoiy"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Namangan"),
				tgbotapi.NewKeyboardButton("Margilon"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Qo'qon"),
				tgbotapi.NewKeyboardButton("Guliston"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Urgench"),
				tgbotapi.NewKeyboardButton("Pop"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("🔙 Back to Main Menu"),
			),
		)
		b.safeSend(msg)

	case "select_to_station":
		// Extract station name from button text (no flag emoji anymore)
		stationName := strings.TrimSpace(text)

		// Debug logging
		log.Printf("DEBUG: Button text: '%s', Extracted destination station: '%s'", text, stationName)

		// Validate station name
		if stationName == "" {
			msg := tgbotapi.NewMessage(chatID, "❌ Invalid station selection. Please try again.")
			b.safeSend(msg)
			return
		}

		userState.ToStation = stationName

		// Check if it's the same station
		if userState.FromStation == userState.ToStation {
			msg := tgbotapi.NewMessage(chatID,
				"❌ Departure and destination stations cannot be the same. Please select a different destination station.")
			msg.ParseMode = "Markdown"
			b.safeSend(msg)
			return
		}

		// Show confirmation and search
		msg := tgbotapi.NewMessage(chatID,
			fmt.Sprintf("✅ *Search Confirmation*\n\n"+
				"🚉 From: *%s*\n"+
				"🎯 To: *%s*\n"+
				"📅 Date: *%s*\n\n"+
				"🔍 Searching for trains...",
				userState.FromStation,
				userState.ToStation,
				userState.SearchDate.Format("2006-01-02")))
		msg.ParseMode = "Markdown"
		b.safeSend(msg)

		// Debug logging before search
		log.Printf("DEBUG: About to search from '%s' to '%s' on %s",
			userState.FromStation, userState.ToStation, userState.SearchDate.Format("2006-01-02"))

		// Perform the search
		b.handleSearchRequest(chatID, userState.FromStation, userState.ToStation, userState.SearchDate)

		// Reset user state after search is complete
		b.resetUserState(chatID)

	default:
		// Unknown step, reset and show main menu
		b.resetUserState(chatID)
		b.handleMainMenuButton(chatID)
	}
}

func (b *Bot) handleSearchTrainsButton(chatID int64) {
	// Reset user state and start station selection
	b.resetUserState(chatID)
	userState := b.getUserState(chatID)
	userState.CurrentStep = "select_from_station"
	userState.SearchDate = time.Now()

	text := `🔍 *Search Trains (Today)*

Please select your departure station:`

	// Create station selection keyboard with all 16 stations
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Toshkent"),
			tgbotapi.NewKeyboardButton("Samarqand"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Buxoro"),
			tgbotapi.NewKeyboardButton("Andijon"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Qarshi"),
			tgbotapi.NewKeyboardButton("Termiz"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Nukus"),
			tgbotapi.NewKeyboardButton("Xiva"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Jizzax"),
			tgbotapi.NewKeyboardButton("Navoiy"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Namangan"),
			tgbotapi.NewKeyboardButton("Margilon"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Qo'qon"),
			tgbotapi.NewKeyboardButton("Guliston"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Urgench"),
			tgbotapi.NewKeyboardButton("Pop"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔙 Back to Main Menu"),
		),
	)
	keyboard.ResizeKeyboard = true
	keyboard.OneTimeKeyboard = false

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard
	b.safeSend(msg)
}

func (b *Bot) handleSearchByDateButton(chatID int64) {
	// Reset user state and start date selection
	b.resetUserState(chatID)
	userState := b.getUserState(chatID)
	userState.CurrentStep = "select_date"
	userState.SearchDate = time.Now().AddDate(0, 0, 1) // Default to tomorrow

	// Show calendar for date selection
	b.showCalendar(chatID, time.Now())
}

// showCalendar displays a calendar for date selection
func (b *Bot) showCalendar(chatID int64, currentDate time.Time) {
	// Get the first day of the month and the number of days
	year, month, _ := currentDate.Date()
	firstDay := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	lastDay := firstDay.AddDate(0, 1, -1)

	// Calculate the day of week for the first day (0 = Sunday, 1 = Monday, etc.)
	// Adjust to make Monday = 0 for better UX
	firstDayWeekday := int(firstDay.Weekday())
	if firstDayWeekday == 0 {
		firstDayWeekday = 7 // Sunday becomes 7
	} else {
		firstDayWeekday-- // Monday becomes 0, Tuesday becomes 1, etc.
	}

	// Create calendar header
	monthNames := []string{
		"January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December",
	}

	calendarText := fmt.Sprintf("📅 Select Travel Date\n\n%s %d\n\n", monthNames[month-1], year)

	// Create calendar grid using the helper function
	keyboard := b.createCalendarGrid(year, month, firstDayWeekday, lastDay.Day())

	// Month navigation row
	prevMonth := currentDate.AddDate(0, -1, 0)
	nextMonth := currentDate.AddDate(0, 1, 0)

	monthRow := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("◀️", fmt.Sprintf("month_%d_%d", prevMonth.Year(), prevMonth.Month())),
		tgbotapi.NewInlineKeyboardButtonData("▶️", fmt.Sprintf("month_%d_%d", nextMonth.Year(), nextMonth.Month())),
	}
	keyboard = append(keyboard, monthRow)

	// Add back button
	backRow := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("🔙 Back to Main Menu", "main_menu"),
	}
	keyboard = append(keyboard, backRow)

	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(keyboard...)

	msg := tgbotapi.NewMessage(chatID, calendarText)
	msg.ReplyMarkup = inlineKeyboard
	b.safeSend(msg)
}

// showCalendarEdit edits an existing calendar message (for month navigation)
func (b *Bot) showCalendarEdit(chatID int64, messageID int, currentDate time.Time) {
	// Get the first day of the month and the number of days
	year, month, _ := currentDate.Date()
	firstDay := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	lastDay := firstDay.AddDate(0, 1, -1)

	// Calculate the day of week for the first day (0 = Sunday, 1 = Monday, etc.)
	// Adjust to make Monday = 0 for better UX
	firstDayWeekday := int(firstDay.Weekday())
	if firstDayWeekday == 0 {
		firstDayWeekday = 7 // Sunday becomes 7
	} else {
		firstDayWeekday-- // Monday becomes 0, Tuesday becomes 1, etc.
	}

	// Create calendar header
	monthNames := []string{
		"January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December",
	}

	calendarText := fmt.Sprintf("📅 Select Travel Date\n\n%s %d\n\n", monthNames[month-1], year)

	// Create calendar grid using the helper function
	keyboard := b.createCalendarGrid(year, month, firstDayWeekday, lastDay.Day())

	// Month navigation row
	prevMonth := currentDate.AddDate(0, -1, 0)
	nextMonth := currentDate.AddDate(0, 1, 0)

	monthRow := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("◀️", fmt.Sprintf("month_%d_%d", prevMonth.Year(), prevMonth.Month())),
		tgbotapi.NewInlineKeyboardButtonData("▶️", fmt.Sprintf("month_%d_%d", nextMonth.Year(), nextMonth.Month())),
	}
	keyboard = append(keyboard, monthRow)

	// Add back button
	backRow := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("🔙 Back to Main Menu", "main_menu"),
	}
	keyboard = append(keyboard, backRow)

	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(keyboard...)

	// Edit the existing message instead of sending a new one
	editMsg := tgbotapi.NewEditMessageText(chatID, messageID, calendarText)
	editMsg.ReplyMarkup = &inlineKeyboard
	b.safeSendEdit(editMsg)
}

// createCalendarGrid creates a properly aligned calendar grid
func (b *Bot) createCalendarGrid(year int, month time.Month, firstDayWeekday int, totalDays int) [][]tgbotapi.InlineKeyboardButton {
	var keyboard [][]tgbotapi.InlineKeyboardButton

	// Add weekday headers row
	weekdayRow := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("Mon", "header"),
		tgbotapi.NewInlineKeyboardButtonData("Tue", "header"),
		tgbotapi.NewInlineKeyboardButtonData("Wed", "header"),
		tgbotapi.NewInlineKeyboardButtonData("Thu", "header"),
		tgbotapi.NewInlineKeyboardButtonData("Fri", "header"),
		tgbotapi.NewInlineKeyboardButtonData("Sat", "header"),
		tgbotapi.NewInlineKeyboardButtonData("Sun", "header"),
	}
	keyboard = append(keyboard, weekdayRow)

	// Calculate total weeks needed for the month
	totalCells := firstDayWeekday + totalDays
	totalWeeks := (totalCells + 6) / 7 // Round up division

	// Create calendar grid with proper cell alignment
	for week := 0; week < totalWeeks; week++ {
		var weekRow []tgbotapi.InlineKeyboardButton

		for dayOfWeek := 0; dayOfWeek < 7; dayOfWeek++ {
			cellIndex := week*7 + dayOfWeek

			if cellIndex < firstDayWeekday || cellIndex >= firstDayWeekday+totalDays {
				// Empty cell (before month starts or after month ends)
				weekRow = append(weekRow, tgbotapi.NewInlineKeyboardButtonData(" ", "empty"))
			} else {
				// Day cell
				dayNumber := cellIndex - firstDayWeekday + 1
				dayText := fmt.Sprintf("%d", dayNumber)
				dateButton := tgbotapi.NewInlineKeyboardButtonData(
					dayText,
					fmt.Sprintf("date_%d_%d_%d", year, month, dayNumber),
				)
				weekRow = append(weekRow, dateButton)
			}
		}

		keyboard = append(keyboard, weekRow)
	}

	return keyboard
}

// handleCalendarCallback handles calendar navigation and date selection
func (b *Bot) handleCalendarCallback(update tgbotapi.Update) {
	callback := update.CallbackQuery
	chatID := callback.Message.Chat.ID
	data := callback.Data

	// Get user state
	userState := b.getUserState(chatID)

	if strings.HasPrefix(data, "month_") {
		// Month navigation - edit the existing message instead of sending a new one
		parts := strings.Split(data, "_")
		if len(parts) == 3 {
			year, _ := strconv.Atoi(parts[1])
			month, _ := strconv.Atoi(parts[2])
			newDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
			b.showCalendarEdit(chatID, callback.Message.MessageID, newDate)
		}
	} else if strings.HasPrefix(data, "date_") {
		// Date selection
		parts := strings.Split(data, "_")
		if len(parts) == 4 {
			year, _ := strconv.Atoi(parts[1])
			month, _ := strconv.Atoi(parts[2])
			day, _ := strconv.Atoi(parts[3])
			selectedDate := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)

			// Check if date is in the past
			if selectedDate.Before(time.Now().Truncate(24 * time.Hour)) {
				msg := tgbotapi.NewMessage(chatID, "❌ Cannot select a date in the past. Please choose a future date.")
				b.safeSend(msg)
				return
			}

			// Store selected date and proceed to station selection
			userState.SearchDate = selectedDate
			userState.CurrentStep = "select_from_station"

			// Show station selection by editing the existing calendar message
			text := fmt.Sprintf("✅ Selected date: %s", selectedDate.Format("2006-01-02"))

			// Create station selection keyboard
			keyboard := tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("Toshkent"),
					tgbotapi.NewKeyboardButton("Samarqand"),
				),
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("Buxoro"),
					tgbotapi.NewKeyboardButton("Andijon"),
				),
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("Qarshi"),
					tgbotapi.NewKeyboardButton("Termiz"),
				),
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("Nukus"),
					tgbotapi.NewKeyboardButton("Xiva"),
				),
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("Jizzax"),
					tgbotapi.NewKeyboardButton("Navoiy"),
				),
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("Namangan"),
					tgbotapi.NewKeyboardButton("Margilon"),
				),
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("Qo'qon"),
					tgbotapi.NewKeyboardButton("Guliston"),
				),
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("Urgench"),
					tgbotapi.NewKeyboardButton("Pop"),
				),
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("🔙 Back to Main Menu"),
				),
			)
			keyboard.ResizeKeyboard = true
			keyboard.OneTimeKeyboard = false

			// Edit the existing calendar message to show just the date confirmation
			editMsg := tgbotapi.NewEditMessageText(chatID, callback.Message.MessageID, text)
			// Remove Markdown parsing to avoid formatting errors
			b.safeSendEdit(editMsg)

			// Send a separate message with the station selection prompt and keyboard
			msg := tgbotapi.NewMessage(chatID, "Please select your departure station:")
			msg.ReplyMarkup = keyboard
			b.safeSend(msg)
		}
	}
}

func (b *Bot) handleViewStationsButton(chatID int64) {
	// Show all 16 stations in a nice format
	response := `🚉 *Available Railway Stations (16 total):*

*Major Cities:*
🇺🇿 **Toshkent** - Capital city
🇺🇿 **Samarqand** - Historic center
🇺🇿 **Buxoro** - Ancient city
🇺🇿 **Andijon** - Eastern hub
🇺🇿 **Qarshi** - Southern center
🇺🇿 **Termiz** - Southern border
🇺🇿 **Nukus** - Karakalpakstan
🇺🇿 **Xiva** - Historic oasis

*Regional Centers:*
🇺🇿 **Jizzax** - Central region
🇺🇿 **Navoiy** - Central mining
🇺🇿 **Namangan** - Fergana Valley
🇺🇿 **Margilon** - Silk city
🇺🇿 **Qo'qon** - Fergana hub
🇺🇿 **Guliston** - Sirdaryo region
🇺🇿 **Urgench** - Khorezm center
🇺🇿 **Pop** - Namangan region

💡 *All stations support train connections!*`

	// Use ReplyKeyboard for consistency
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔙 Back to Main Menu"),
		),
	)
	keyboard.ResizeKeyboard = true
	keyboard.OneTimeKeyboard = false

	msg := tgbotapi.NewMessage(chatID, response)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard
	b.safeSend(msg)
}

func (b *Bot) handleChangeLanguageButton(chatID int64) {
	text := `🌍 *Change Language*

Choose your preferred language for the bot interface:`

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🇺🇿 O'zbekcha"),
			tgbotapi.NewKeyboardButton("🇷🇺 Русский"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🇺🇸 English"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔙 Back to Main Menu"),
		),
	)
	keyboard.ResizeKeyboard = true
	keyboard.OneTimeKeyboard = false

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard
	b.safeSend(msg)
}

func (b *Bot) handleHelpButton(chatID int64) {
	helpText := `🚂 *ChiptaTop Train Bot Help*

🔍 *How to Use:*
• Use the buttons below to navigate
• Search for trains between any stations
• View available dates and times
• Change language as needed

📋 *Available Options:*
• Search Trains - Find trains for today
• Search by Date - Find trains for specific date
• View Stations - See all available stations
• Change Language - Switch between Uzbek/Russian/English

💡 *Tips:*
• All major cities are supported
• Results show available seats and prices
• Automatic language detection`

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔙 Back to Main Menu"),
		),
	)
	keyboard.ResizeKeyboard = true
	keyboard.OneTimeKeyboard = false

	msg := tgbotapi.NewMessage(chatID, helpText)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard
	b.safeSend(msg)
}

func (b *Bot) handleMainMenuButton(chatID int64) {
	// Send the main menu again
	b.handleStartCommand(tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: chatID},
		},
	})
}

func (b *Bot) handleSearchRequest(chatID int64, from, to string, date time.Time) {
	// Send "searching" message
	searchingMsg := tgbotapi.NewMessage(chatID,
		fmt.Sprintf("🔍 Searching trains from %s to %s on %s...",
			from, to, date.Format("2006-01-02")))
	// Remove Markdown parsing to avoid formatting errors
	b.safeSend(searchingMsg)

	// Perform search with retry logic
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	searchParams := train.TrainSearchParams{
		From: from,
		To:   to,
		Date: date,
	}

	// Get all trains from API response with retry logic
	response, err := b.searchTrainsWithRetry(ctx, searchParams)
	if err != nil {
		log.Printf("Train search error after retries: %v", err)

		var errorMsg string
		if strings.Contains(err.Error(), "403") || strings.Contains(err.Error(), "CSRF") {
			errorMsg = "❌ Authentication Error\n\n" +
				"Unable to authenticate with railway service. Please try again later.\n\n" +
				"If this problem persists, the railway service may be temporarily unavailable."
		} else if strings.Contains(err.Error(), "failed to search trains") {
			errorMsg = "❌ Search Failed\n\n" +
				"Could not connect to railway service after multiple attempts. This might be because:\n" +
				"• Network connection issues\n" +
				"• Railway service is temporarily unavailable\n" +
				"• High server load\n\n" +
				"Please try again in a few moments."
		} else {
			errorMsg = "❌ Search Error\n\n" +
				"An unexpected error occurred while searching for trains. Please try again later."
		}

		msg := tgbotapi.NewMessage(chatID, errorMsg)
		b.safeSend(msg)
		return
	}

	// Extract all trains from response
	var trains []train.Train
	if response.Data != nil && response.Data.Directions.Forward != nil {
		trains = response.Data.Directions.Forward.Trains
	}

	// Format and send results
	if len(trains) == 0 {
		// Reset user state since search is complete
		b.resetUserState(chatID)

		msg := tgbotapi.NewMessage(chatID,
			fmt.Sprintf("❌ No available trains found from *%s* to *%s* on *%s*.\n\n"+
				"Try:\n• Different dates\n• Alternative station names\n• Use the View Stations button to see available stations",
				from, to, date.Format("2006-01-02")))
		msg.ParseMode = "Markdown"

		// Send main menu
		keyboard := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("🔍 Search Trains"),
				tgbotapi.NewKeyboardButton("📅 Search by Date"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("🚉 View Stations"),
				tgbotapi.NewKeyboardButton("🌍 Change Language"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("❓ Help"),
			),
		)
		keyboard.ResizeKeyboard = true
		keyboard.OneTimeKeyboard = false
		msg.ReplyMarkup = keyboard

		b.safeSend(msg)
		return
	}

	// Format search results
	results := b.trainService.FormatSearchResults(trains)

	// Send results with main menu
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔍 Search Trains"),
			tgbotapi.NewKeyboardButton("📅 Search by Date"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🚉 View Stations"),
			tgbotapi.NewKeyboardButton("🌍 Change Language"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("❓ Help"),
		),
	)
	keyboard.ResizeKeyboard = true
	keyboard.OneTimeKeyboard = false

	msg := tgbotapi.NewMessage(chatID, results)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard
	b.safeSend(msg)
}

func (b *Bot) handleLanguageChange(chatID int64, language string) {
	// Change the train service language
	b.trainService.SetLanguage(language)

	var text string
	switch language {
	case "uz":
		text = "🇺🇿 *Til o'zgartirildi!*\n\nO'zbek tiliga o'tkazildi. Endi barcha API so'rovlari o'zbek tilida bo'ladi."
	case "ru":
		text = "🇷🇺 *Язык изменен!*\n\nПереключено на русский язык. Теперь все API запросы будут на русском языке."
	case "en":
		text = "🇺🇸 *Language changed!*\n\nSwitched to English. Now all API requests will be in English."
	default:
		text = "❌ Unknown language"
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Back to Main Menu", "main_menu"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard
	b.safeSend(msg)
}

func (b *Bot) handleStationsCommand(update tgbotapi.Update) {
	stations := b.trainService.GetStationSuggestions("")

	var response strings.Builder
	response.WriteString("🚉 *Available Railway Stations:*\n\n")

	for i, station := range stations {
		response.WriteString(fmt.Sprintf("• %s", station))
		if i < len(stations)-1 {
			response.WriteString("\n")
		}
	}

	response.WriteString("\n\n💡 Use these names in your search requests!")

	// Add back button
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔙 Back to Main Menu"),
		),
	)
	keyboard.ResizeKeyboard = true
	keyboard.OneTimeKeyboard = false

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, response.String())
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard
	b.safeSend(msg)
}

func (b *Bot) handleSearchCommand(update tgbotapi.Update) {
	args := strings.Fields(update.Message.CommandArguments())
	if len(args) < 2 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			"❌ Please provide departure and arrival stations.\n\n"+
				"Example: `/search Toshkent Samarqand`")
		msg.ParseMode = "Markdown"
		b.safeSend(msg)
		return
	}

	from := args[0]
	to := args[1]
	date := time.Now()

	b.performTrainSearch(update.Message.Chat.ID, from, to, date)
}

func (b *Bot) handleSearchDateCommand(update tgbotapi.Update) {
	args := strings.Fields(update.Message.CommandArguments())
	if len(args) < 3 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			"❌ Please provide departure, arrival stations and date.\n\n"+
				"Example: `/search_date Toshkent Samarqand 2025-01-15`")
		msg.ParseMode = "Markdown"
		b.safeSend(msg)
		return
	}

	from := args[0]
	to := args[1]
	dateStr := args[2]

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			"❌ Invalid date format. Please use YYYY-MM-DD format.\n\n"+
				"Example: `2025-01-15`")
		msg.ParseMode = "Markdown"
		b.safeSend(msg)
		return
	}

	b.performTrainSearch(update.Message.Chat.ID, from, to, date)
}

func (b *Bot) performTrainSearch(chatID int64, from, to string, date time.Time) {
	// Send "searching" message
	searchingMsg := tgbotapi.NewMessage(chatID,
		fmt.Sprintf("🔍 Searching trains from %s to %s on %s...",
			from, to, date.Format("2006-01-02")))
	// Remove Markdown parsing to avoid formatting errors
	b.safeSend(searchingMsg)

	// Perform search with retry logic
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	searchParams := train.TrainSearchParams{
		From: from,
		To:   to,
		Date: date,
	}

	trains, err := b.findAvailableTrainsWithRetry(ctx, searchParams)
	if err != nil {
		log.Printf("Train search error after retries: %v", err)

		var errorMsg string
		if strings.Contains(err.Error(), "403") || strings.Contains(err.Error(), "CSRF") {
			errorMsg = "❌ Authentication Error\n\n" +
				"Unable to authenticate with railway service. Please try again later.\n\n" +
				"If this problem persists, the railway service may be temporarily unavailable."
		} else if strings.Contains(err.Error(), "failed to search trains") {
			errorMsg = "❌ Search Failed\n\n" +
				"Could not connect to railway service after multiple attempts. This might be because:\n" +
				"• Network connection issues\n" +
				"• Railway service is temporarily unavailable\n" +
				"• High server load\n\n" +
				"Please try again in a few moments."
		} else {
			errorMsg = "❌ Search Error\n\n" +
				"An unexpected error occurred while searching for trains. Please try again later."
		}

		msg := tgbotapi.NewMessage(chatID, errorMsg)
		b.safeSend(msg)
		return
	}

	// Format and send results
	if len(trains) == 0 {
		msg := tgbotapi.NewMessage(chatID,
			fmt.Sprintf("❌ No available trains found from *%s* to *%s* on *%s*.\n\n"+
				"Try:\n• Different dates\n• Alternative station names\n• Use /stations to see available stations",
				from, to, date.Format("2006-01-02")))
		msg.ParseMode = "Markdown"
		b.safeSend(msg)
		return
	}

	// Send results (split if too long)
	results := b.trainService.FormatSearchResults(trains)
	b.sendLongMessage(chatID, results)
}

func (b *Bot) sendLongMessage(chatID int64, text string) {
	const maxLength = 4096

	if len(text) <= maxLength {
		msg := tgbotapi.NewMessage(chatID, text)
		msg.ParseMode = "Markdown"
		b.safeSend(msg)
		return
	}

	// Split long messages
	parts := b.splitMessage(text, maxLength)
	for i, part := range parts {
		msg := tgbotapi.NewMessage(chatID, part)
		msg.ParseMode = "Markdown"

		if i > 0 {
			time.Sleep(500 * time.Millisecond) // Small delay between messages
		}

		b.safeSend(msg)
	}
}

func (b *Bot) splitMessage(text string, maxLength int) []string {
	if len(text) <= maxLength {
		return []string{text}
	}

	var parts []string
	lines := strings.Split(text, "\n")
	var currentPart strings.Builder

	for _, line := range lines {
		if currentPart.Len()+len(line)+1 > maxLength {
			if currentPart.Len() > 0 {
				parts = append(parts, currentPart.String())
				currentPart.Reset()
			}
		}

		if currentPart.Len() > 0 {
			currentPart.WriteString("\n")
		}
		currentPart.WriteString(line)
	}

	if currentPart.Len() > 0 {
		parts = append(parts, currentPart.String())
	}

	return parts
}

func (b *Bot) safeSend(msg tgbotapi.MessageConfig) {
	if _, err := b.api.Send(msg); err != nil {
		log.Printf("send error: %v", err)
		time.Sleep(200 * time.Millisecond)
	}
}

func (b *Bot) safeSendEdit(msg tgbotapi.EditMessageTextConfig) {
	if _, err := b.api.Send(msg); err != nil {
		log.Printf("edit message error: %v", err)
		time.Sleep(200 * time.Millisecond)
	}
}

// searchTrainsWithRetry performs train search with automatic retry logic
func (b *Bot) searchTrainsWithRetry(ctx context.Context, params train.TrainSearchParams) (*train.SearchTrainsResponse, error) {
	const maxRetries = 3
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Log retry attempt
		if attempt > 1 {
			log.Printf("Retrying train search (attempt %d/%d)...", attempt, maxRetries)
		}

		// Perform the search
		response, err := b.trainService.SearchTrains(ctx, params)
		if err == nil {
			// Success - return the response
			if attempt > 1 {
				log.Printf("Train search succeeded on attempt %d", attempt)
			}
			return response, nil
		}

		lastErr = err

		// Don't retry on authentication errors (403/CSRF) - these won't be fixed by retrying
		if strings.Contains(err.Error(), "403") || strings.Contains(err.Error(), "CSRF") {
			log.Printf("Authentication error, not retrying: %v", err)
			break
		}

		// Don't retry on context cancellation
		if ctx.Err() != nil {
			log.Printf("Context cancelled, not retrying: %v", ctx.Err())
			break
		}

		// If this is the last attempt, don't sleep
		if attempt == maxRetries {
			break
		}

		// Calculate delay with exponential backoff: 1s, 2s, 4s
		delay := time.Duration(attempt) * time.Second
		log.Printf("Search failed (attempt %d/%d), retrying in %v: %v", attempt, maxRetries, delay, err)

		// Wait before retrying
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
			continue
		}
	}

	return nil, fmt.Errorf("failed to search trains after %d attempts: %w", maxRetries, lastErr)
}

// findAvailableTrainsWithRetry performs available trains search with automatic retry logic
func (b *Bot) findAvailableTrainsWithRetry(ctx context.Context, params train.TrainSearchParams) ([]train.Train, error) {
	const maxRetries = 3
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Log retry attempt
		if attempt > 1 {
			log.Printf("Retrying available trains search (attempt %d/%d)...", attempt, maxRetries)
		}

		// Perform the search
		trains, err := b.trainService.FindAvailableTrains(ctx, params)
		if err == nil {
			// Success - return the trains
			if attempt > 1 {
				log.Printf("Available trains search succeeded on attempt %d", attempt)
			}
			return trains, nil
		}

		lastErr = err

		// Don't retry on authentication errors (403/CSRF) - these won't be fixed by retrying
		if strings.Contains(err.Error(), "403") || strings.Contains(err.Error(), "CSRF") {
			log.Printf("Authentication error, not retrying: %v", err)
			break
		}

		// Don't retry on context cancellation
		if ctx.Err() != nil {
			log.Printf("Context cancelled, not retrying: %v", ctx.Err())
			break
		}

		// If this is the last attempt, don't sleep
		if attempt == maxRetries {
			break
		}

		// Calculate delay with exponential backoff: 1s, 2s, 4s
		delay := time.Duration(attempt) * time.Second
		log.Printf("Available trains search failed (attempt %d/%d), retrying in %v: %v", attempt, maxRetries, delay, err)

		// Wait before retrying
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
			continue
		}
	}

	return nil, fmt.Errorf("failed to find available trains after %d attempts: %w", maxRetries, lastErr)
}
