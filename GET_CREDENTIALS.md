# 🔑 How to Get Railway.uz API Credentials

## 🚀 Quick Method: Extract from Your Working Curl

Since you already have a working curl command, let's extract the credentials from it:

### Your Curl Command:
```bash
curl 'https://eticket.railway.uz/api/v3/handbook/trains/list' \
  -H 'X-XSRF-TOKEN: 003c204f-01ca-4820-85cc-925fa66a6c41' \
  -H 'Cookie: _ga=GA1.1.2112044518.1734322567; __stripe_mid=aa188b56-afd0-4e79-9731-b08f6971b6d837b937; G_ENABLED_IDPS=google; _ga_K4H2SZ7MWK=GS2.1.s1756394922$o5$g1$t1756395266$j60$l0$h0; XSRF-TOKEN=003c204f-01ca-4820-85cc-925fa66a6c41; _ga_R5LGX7P1YR=GS2.1.s1756628889$o5$g0$t1756628889$j60$l0$h0; __stripe_sid=c4df2400-9acb-4409-8f74-94444c8560ef7e8757'
```

### Extract These Values:

1. **RAILWAY_XSRF_TOKEN**: 
   ```
   003c204f-01ca-4820-85cc-925fa66a6c41
   ```

2. **RAILWAY_COOKIES**:
   ```
   _ga=GA1.1.2112044518.1734322567; __stripe_mid=aa188b56-afd0-4e79-9731-b08f6971b6d837b937; G_ENABLED_IDPS=google; _ga_K4H2SZ7MWK=GS2.1.s1756394922$o5$g1$t1756395266$j60$l0$h0; XSRF-TOKEN=003c204f-01ca-4820-85cc-925fa66a6c41; _ga_R5LGX7P1YR=GS2.1.s1756628889$o5$g0$t1756628889$j60$l0$h0; __stripe_sid=c4df2400-9acb-4409-8f74-94444c8560ef7e8757
   ```

## 📝 Update Your .env File

Add these lines to your `.env` file:

```bash
# Telegram Bot Configuration
TELEGRAM_BOT_TOKEN=your_existing_token

# Environment
ENVIRONMENT=production

# Railway.uz API Authentication (from your curl command)
RAILWAY_XSRF_TOKEN=003c204f-01ca-4820-85cc-925fa66a6c41
RAILWAY_COOKIES=_ga=GA1.1.2112044518.1734322567; __stripe_mid=aa188b56-afd0-4e79-9731-b08f6971b6d837b937; G_ENABLED_IDPS=google; _ga_K4H2SZ7MWK=GS2.1.s1756394922$o5$g1$t1756395266$j60$l0$h0; XSRF-TOKEN=003c204f-01ca-4820-85cc-925fa66a6c41; _ga_R5LGX7P1YR=GS2.1.s1756628889$o5$g0$t1756628889$j60$l0$h0; __stripe_sid=c4df2400-9acb-4409-8f74-94444c8560ef7e8757
```

## 🚀 Test Production Mode

1. **Start the bot:**
   ```bash
   make run
   ```

2. **Expected startup logs:**
   ```
   Railway API authentication configured
   Railway API authentication configured for production
   Bot @YourBotName started in production environment
   ```

3. **Test in Telegram:**
   ```
   /search Toshkent Samarqand
   ```

4. **Expected result:** Real train data instead of error messages!

## 🔄 If Credentials Expire

XSRF tokens typically expire after 24-48 hours. If you start getting 403 errors again:

### Method 1: Refresh Browser Session
1. Go to https://eticket.railway.uz
2. Clear cookies and refresh
3. Search for trains
4. Extract new credentials using browser dev tools

### Method 2: Use Browser Developer Tools
1. Open https://eticket.railway.uz
2. Open Developer Tools (F12)
3. Go to **Network** tab
4. Search for any train route
5. Find the `/api/v3/handbook/trains/list` request
6. Copy new `X-XSRF-TOKEN` and `Cookie` headers
7. Update your `.env` file

## ⚠️ Important Notes

- **Credentials are session-based** and will expire
- **Don't share credentials** publicly
- **Monitor for 403 errors** which indicate expired tokens
- **Test your curl command first** to ensure credentials work

## 🎉 Ready for Production!

Once you add the credentials, your bot will:
- ✅ Return real train schedules
- ✅ Show actual seat availability  
- ✅ Display current prices in UZS
- ✅ Work with all 16 railway stations
- ✅ Handle multiple trains per route

Your bot is now ready for production use with live railway data! 🚂✨
