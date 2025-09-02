# 🚂 ChiptaTop Production Setup Guide

## 🚀 Quick Production Setup

### 1. Get Railway.uz API Credentials

To get real train data, you need to extract authentication credentials from railway.uz:

#### Method 1: Browser Developer Tools
1. Open https://eticket.railway.uz in Chrome/Firefox
2. Open Developer Tools (F12)
3. Go to **Network** tab
4. Search for trains (any route)
5. Find the API request to `/api/v3/handbook/trains/list`
6. Copy the **Request Headers**:
   - `X-XSRF-TOKEN`: Copy the token value
   - `Cookie`: Copy the entire cookie string

#### Method 2: Using Your Curl Command
From your working curl command, extract:
```bash
curl 'https://eticket.railway.uz/api/v3/handbook/trains/list' \
  -H 'X-XSRF-TOKEN: 003c204f-01ca-4820-85cc-925fa66a6c41' \
  -H 'Cookie: _ga=GA1.1.2112044518.1734322567; __stripe_mid=...'
```

**Extract these values:**
- `RAILWAY_XSRF_TOKEN`: `003c204f-01ca-4820-85cc-925fa66a6c41`
- `RAILWAY_COOKIES`: `_ga=GA1.1.2112044518.1734322567; __stripe_mid=...`

### 2. Configure Environment Variables

Create `.env` file in project root:

```bash
# Telegram Bot Configuration
TELEGRAM_BOT_TOKEN=your_telegram_bot_token

# Environment
ENVIRONMENT=production

# Railway.uz API Authentication
RAILWAY_XSRF_TOKEN=your_xsrf_token_here
RAILWAY_COOKIES=your_full_cookie_string_here
```

### 3. Example Production .env File

```bash
# Telegram Bot
TELEGRAM_BOT_TOKEN=1234567890:AAEhBOweik6ad6PsVMRxjeQfq1_lbJrGhoc

# Environment  
ENVIRONMENT=production

# Railway API (example values - use your real ones)
RAILWAY_XSRF_TOKEN=003c204f-01ca-4820-85cc-925fa66a6c41
RAILWAY_COOKIES=_ga=GA1.1.2112044518.1734322567; __stripe_mid=aa188b56-afd0-4e79-9731-b08f6971b6d837b937; G_ENABLED_IDPS=google; _ga_K4H2SZ7MWK=GS2.1.s1756394922$o5$g1$t1756395266$j60$l0$h0; XSRF-TOKEN=003c204f-01ca-4820-85cc-925fa66a6c41
```

### 4. Run in Production Mode

```bash
# Install dependencies
go mod tidy

# Run the bot
make run
# or
go run ./cmd/bot
```

**Expected startup logs:**
```
Railway API authentication configured
Railway API authentication configured for production  
Bot @YourBotName started in production environment
```

## 🧪 Testing Production API

### Test Commands in Telegram:

```bash
# Basic search (should return real trains)
/search Toshkent Samarqand

# Date-specific search  
/search_date Toshkent Buxoro 2025-09-02

# Different routes
/search Andijon Termiz
/search Nukus Urgench
```

### Expected Production Response:
```
🎫 Found 12 available train(s):

🚂 *Afrosiyob* (778Ф)
📍 TOSHKENT → SAMARQAND
🕐 06:03 - 08:21 (02:18)
📅 02.09.2025
🚄 Route: Toshkent Markaziy → Buxoro

💺 *Available seats:*
*O'rindiqli* (77 seats):
  • 1В: 11 seats - 545 000 UZS
  • 2Е: 66 seats - 270 000 UZS

────────────────────────────────

🚂 *Sharq* (710Ф)
📍 TOSHKENT → SAMARQAND
🕐 08:40 - 12:05 (03:25)
📅 02.09.2025
🚄 Route: Toshkent Markaziy → Buxoro

💺 *Available seats:*
*O'rindiqli* (248 seats):
  • 1С: 154 seats - 266 700 UZS
  • 1В: 25 seats - 498 030 UZS
  • 2В: 69 seats - 179 250 UZS
```

## 🔧 Troubleshooting

### Authentication Issues

**Problem**: Still getting 403 errors
**Solutions**:
1. **Refresh credentials**: XSRF tokens expire, get new ones
2. **Check cookie format**: Ensure full cookie string is copied
3. **Test with curl**: Verify credentials work with original curl command

**Problem**: "Railway API authentication not configured"
**Solutions**:
1. Check `.env` file exists in project root
2. Verify environment variable names are exact
3. Restart the bot after adding credentials

### API Rate Limits

**Problem**: Getting rate limited
**Solutions**:
1. Add delays between requests (already implemented)
2. Use different IP if needed
3. Respect railway.uz terms of service

### Token Expiration

**Problem**: Credentials stop working after some time
**Solutions**:
1. XSRF tokens typically expire after 24-48 hours
2. Set up a refresh mechanism or manual renewal
3. Monitor logs for authentication failures

## 🔒 Security Best Practices

### Environment Variables
- ✅ Never commit `.env` to git
- ✅ Use different tokens for different environments
- ✅ Rotate credentials regularly
- ✅ Monitor for unauthorized usage

### Production Deployment
- ✅ Use environment-specific configs
- ✅ Set up proper logging
- ✅ Monitor API usage and errors
- ✅ Implement graceful error handling

## 📊 Monitoring

### Key Metrics to Monitor:
- **API Success Rate**: Should be >95%
- **Response Times**: Typical <5 seconds
- **Error Patterns**: Watch for authentication failures
- **User Activity**: Track popular routes

### Log Messages to Watch:
```
✅ "Railway API authentication configured"
❌ "Train search error: failed to search trains"
❌ "API request failed with status 403"
```

## 🚀 Deployment Options

### Local Development
```bash
make run
```

### Docker Production
```bash
# Build
make docker-build

# Run with environment variables
docker run --rm \
  -e TELEGRAM_BOT_TOKEN=your_token \
  -e RAILWAY_XSRF_TOKEN=your_xsrf \
  -e RAILWAY_COOKIES=your_cookies \
  -e ENVIRONMENT=production \
  chiptatop-bot:dev
```

### Cloud Deployment
- Set environment variables in your cloud platform
- Ensure secure credential management
- Set up monitoring and logging

## ✅ Production Checklist

- [ ] Telegram bot token configured
- [ ] Railway XSRF token obtained
- [ ] Railway cookies extracted
- [ ] `.env` file created with all credentials
- [ ] Bot starts without errors
- [ ] Authentication logs show "configured"
- [ ] Test searches return real train data
- [ ] Error handling works for edge cases
- [ ] Monitoring/logging set up

Your bot is now ready for production use with real railway.uz data! 🎉
