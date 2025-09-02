# 🚂 ChiptaTop Telegram Bot Testing Guide

## 🚀 Quick Start

### 1. Setup Your Bot

First, create a Telegram bot and get your token:

1. Open Telegram and search for `@BotFather`
2. Send `/newbot` and follow instructions
3. Copy your bot token
4. Create `.env` file in project root:

```bash
TELEGRAM_BOT_TOKEN=your_bot_token_here
ENVIRONMENT=development
```

### 2. Run the Bot

```bash
# Install dependencies
go mod tidy

# Run the bot
make run
# or
go run ./cmd/bot
```

You should see:
```
Bot @YourBotName started in development environment
```

## 🧪 Testing Commands

### Basic Commands

1. **Start the bot:**
   ```
   /start
   ```
   Expected: Welcome message with quick start guide

2. **Get help:**
   ```
   /help
   ```
   Expected: Complete command list with examples

3. **View stations:**
   ```
   /stations
   ```
   Expected: List of all 16 available railway stations

### Train Search Commands

#### 1. Basic Search (Today's Date)
```
/search Toshkent Samarqand
```

**Expected behavior:**
- Bot shows "🔍 Searching trains..." message
- Then shows either:
  - Train results with times, prices, available seats
  - "Search Failed" message (if no API auth)

#### 2. Date-Specific Search
```
/search_date Toshkent Buxoro 2025-09-02
```

**Expected behavior:**
- Same as basic search but for specific date
- Date validation (rejects invalid formats)

#### 3. Test Different Stations
```
/search Andijon Termiz
/search Nukus Urgench
/search Xiva Qarshi
```

#### 4. Test Alternative Spellings
```
/search Tashkent Samarkand
/search Bukhara Khiva
```

### Error Handling Tests

#### 1. Invalid Commands
```
/search
/search Toshkent
/search_date Toshkent Samarqand
/search_date Toshkent Samarqand invalid-date
```

Expected: Helpful error messages with examples

#### 2. Unknown Stations
```
/search UnknownCity AnotherCity
```

Expected: Station code returned as-is (will fail API call gracefully)

## 📱 Expected Bot Responses

### 1. Successful Search Response Format:
```
🎫 Found X available train(s):

🚂 *Afrosiyob* (778Ф)
📍 TOSHKENT → SAMARQAND  
🕐 06:03 - 08:21 (02:18)
📅 02.09.2025
🚄 Route: Toshkent Markaziy → Buxoro

💺 *Available seats:*
*O'rindiqli* (77 seats):
  • 1В: 11 seats - 545 000 UZS
  • 2Е: 66 seats - 270 000 UZS
```

### 2. No Authentication Response:
```
❌ Search Failed

Could not connect to railway service. This might be because:
• No authentication credentials set
• Network connection issues  
• Railway service is temporarily unavailable

💡 Note: This is a demo version. In production, you would need proper API authentication.
```

### 3. Stations List Response:
```
🚉 Available Railway Stations:

• Andijon
• Buxoro
• Guliston
• Jizzax
• Margilon
• Namangan
• Navoiy
• Nukus
• Pop
• Qarshi
• Qoqon
• Samarqand
• Termiz
• Toshkent
• Urgench
• Xiva

💡 Use these names in your search commands!
```

## 🔐 Adding Real API Authentication (Optional)

To test with real railway.uz API:

1. Get authentication credentials from railway.uz (XSRF token and cookies)
2. Add to your bot initialization in `internal/bot/bot.go`:

```go
func New(cfg config.Config) (*Bot, error) {
    // ... existing code ...
    
    trainService := train.NewService()
    
    // Add real authentication
    trainService.SetAuthCredentials(
        "your_xsrf_token_here",
        "your_cookies_here",
    )
    
    // ... rest of code ...
}
```

## 🐛 Troubleshooting

### Bot Not Starting
- Check `.env` file exists with valid `TELEGRAM_BOT_TOKEN`
- Verify bot token is correct (starts with numbers, ends with letters)
- Check internet connection

### Commands Not Working
- Make sure commands start with `/`
- Check for typos in command names
- Try `/help` to see available commands

### Search Always Fails
- Expected behavior without API authentication
- Bot will show helpful error message
- Station code mapping still works correctly

### Long Response Messages
- Bot automatically splits long messages (>4096 chars)
- Multiple messages sent with small delays

## 📊 Testing Checklist

- [ ] Bot starts successfully
- [ ] `/start` command works
- [ ] `/help` shows all commands
- [ ] `/stations` lists all 16 stations
- [ ] `/search` validates arguments
- [ ] `/search_date` validates date format
- [ ] Error messages are helpful
- [ ] Station name mapping works (Tashkent → Toshkent)
- [ ] Alternative spellings accepted
- [ ] Long messages split properly
- [ ] Bot handles network errors gracefully

## 🎯 Demo Mode vs Production

**Demo Mode (No Auth):**
- Station mapping works ✅
- Command parsing works ✅
- Error handling works ✅
- API calls fail gracefully ✅
- Shows helpful "demo" message ✅

**Production Mode (With Auth):**
- All demo features ✅
- Real train data from railway.uz ✅
- Live ticket availability ✅
- Real-time pricing ✅

The bot is fully functional in demo mode and ready for production with proper API credentials!
