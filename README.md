# Webscrapers
This contains a set of webscrapers I personally use.  

They make use of [playwright-go](https://github.com/playwright-community/playwright-go) so that a single binary only is needed - browser depencies are auto downloaded.

* [Aviva](#user-content-aviva)
* [Aviva My Money](#user-content-aviva-my-money)
* [Fund](#user-content-fund)
* [Home Assistant Screen Shot](#user-content-home-assistant-screen-shot)
* [Home Assistant Screen Shot Addon](#user-content-home-assistant-screen-shot-addon)
* [Moneyfarm](#user-content-moneyfarm)
* [Money Hub](#user-content-money-hub)
* [Nutmeg](#user-content-nutmeg)
* [Octopus Wheel](#user-content-octopus-wheel)
* [Desktop Testing](#user-content-desktop-testing)
* [Pension Scripting](#user-content-pension-scripting)

## [Aviva](aviva)

Connect to [Aviva](https://www.direct.aviva.co.uk/MyAccount/login), login, process one-time-password and return account balance.

```
Retrive Aviva balance via web scraping

Usage:
  bin/aviva [options]

Options:
  -headless
    	Headless mode (default true)
  -otpcleancommand string
    	Command to clean previous one time password
  -otpcommand string
    	Command to get one time password
  -otppath string
    	Path to file containing one time password message (default "otp/aviva")
  -password string
    	Aviva password
  -username string
    	Aviva username

Environment variables:
  $HEADLESS - Headless mode
  $OTP_COMMAND - Command to get one time password
  $OTP_PATH - Path to file containing one time password message
  $AVIVA_USERNAME - Aviva username
  $AVIVA_PASSWORD - Aviva password
```

## [Aviva My Money](avivamymoney)

Connect to [Aviva My Money](https://www.avivamymoney.co.uk/Login), login and return account balance.

```
Retrive Aviva my money balance via web scraping

Usage:
  bin/avivamymoney [options]

Options:
  -headless
    	Headless mode (default true)
  -password string
    	Aviva my money password
  -username string
    	Aviva my money username
  -word string
    	Aviva my money memorable word

Environment variables:
  $HEADLESS - Headless mode
  $AVIVAMYMONEY_USERNAME - Aviva my money username
  $AVIVAMYMONEY_PASSWORD - Aviva my money password
  $AVIVAMYMONEY_WORD - Aviva my money memorable word
```

## [Fund](fund)

Connect to [Financial Times](https://markets.ft.com) and return fund value.

```
Retrive fund value via web scraping

Usage:
  bin/fund [options]

Options:
  -fund string
    	Fund name
  -headless
    	Headless mode (default true)

Environment variables:
  $HEADLESS - Headless mode
  $FUND - Fund name
```

## [Home Assistant Screen Shot](ha_ss)

Connect to local Home Assistant instance, login and take a screen shot based on CSS selector.

```
Connect to Home Assistant and take a screenshot by CSS selector

Usage:
  bin/ha_ss [options]

Options:
  -css string
    	Home assistant CSS selector
  -headless
    	Headless mode (default true)
  -password string
    	Home assistant password
  -path string
    	Output screenshot path (default "output.png")
  -restport int
    	If set, startup REST server at given port
  -url string
    	Home assistant page URL
  -username string
    	Home assistant username

Environment variables:
  $HEADLESS - Headless mode
  $HA_CSS - Home assistant CSS selector
  $HA_PATH - Output screenshot path
  $HA_USERNAME - Home assistant username
  $HA_PASSWORD - Home assistant password
  $HA_RESTPORT - If set, startup REST server at given port
  $HA_URL - Home assistant page URL
```

## [Home Assistant Screen Shot Addon](ha_ss_addon)

Simple Home Assistant addon that hosts ha_ss as a REST server.  This enables automations to include charges with notifications
such as :

```
alias: Car - next green time to charge
description: ""
triggers:
  - entity_id:
      - binary_sensor.octopus_energy_a_xxxxxxx_greenness_forecast_highlighted
    attribute: next_start
    trigger: state
conditions: []
actions:
  - action: rest_command.screenshot
    data:
      url: http://homeassistant.local:8123/lovelace/overview
      css: >-
        div.card:nth-child(4) > hui-card:nth-child(1) >
        octopus-energy-greenness-forecast-card:nth-child(1)
      filename: greenness-forecast.png
  - action: notify.gaselectricity
    data:
      message: >-
        Next octopus greenness forcast is {{
        state_attr('binary_sensor.octopus_energy_a_xxxxxxx_greenness_forecast_highlighted',
        'next_start')|as_local()|as_timestamp()|timestamp_custom('%a %b %-d,
        %I:%M %p') }} to {{
        state_attr('binary_sensor.octopus_energy_a_xxxxxxx_greenness_forecast_highlighted',
        'next_end')|as_local()|as_timestamp()|timestamp_custom('%a %b %-d, %I:%M
        %p') }} - see https://octopus.energy/smart/greener-days
      data:
        attachments:
          - /config/screenshots/greenness-forecast.png
mode: single
```

The steps to install this addon are (currently) :

1. Copy Dockerfile* *.yaml run.sh ha_ss-linux-arm64 to the home assistant /addons/ha_ss_addon directory
2. Refresh addons on Home Assistant
3. In the Addon Store, select the local screenshot tool addon and install
4. In the new addon, build the image and start

configuration.yaml will need :

```
rest_command:
  screenshot:
    url: http://localhost:3500
    method: post
    headers:
      accept: "application/json, text/html"
    payload: '{"url": "{{ url }}", "css": "{{ css }}", "filename": "{{ filename }}"}'
    content_type:  'application/json; charset=utf-8'
    timeout: 120
```

(I'm pretty sure there is a better way to communicate between ha_ss automations)

## [Moneyfarm](moneyfarm)

Connect to [Moneyfarm](hhttps://app.moneyfarm.com/gb/sign-in), login, process one-time-password and return account balance.

```
Retrive Moneyfarm balance via web scraping

Usage:
  bin/moneyfarm [options]

Options:
  -headless
    	Headless mode (default true)
  -otpcleancommand string
    	Command to clean previous one time password
  -otpcommand string
    	Command to get one time password
  -otppath string
    	Path to file containing one time password message (default "otp/moneyfarm")
  -password string
    	Moneyfarm password
  -username string
    	Moneyfarm username

Environment variables:
  $HEADLESS - Headless mode
  $OTP_CLEANCOMMAND - Command to clean previous one time password
  $OTP_COMMAND - Command to get one time password
  $OTP_PATH - Path to file containing one time password message
  $MONEYFARM_USERNAME - Moneyfarm username
  $MONEYFARM_PASSWORD - Moneyfarm password
```

## [Money Hub](moneyhub)

Connect to [Money Hub](https://client.moneyhub.co.uk), login and update the balance of one asset.

```
Update Moneyhub balance via web scraping

Usage:
  bin/moneyhub [options]

Options:
  -account string
    	Moneyhub account
  -balance float
    	Moneyhub balance for the account
  -headless
    	Headless mode (default true)
  -password string
    	Moneyhub password
  -username string
    	Moneyhub username

Environment variables:
  $HEADLESS - Headless mode
  $MONEYHUB_USERNAME - Moneyhub username
  $MONEYHUB_PASSWORD - Moneyhub password
  $MONEYHUB_ACCOUNT - Moneyhub account
  $MONEYHUB_BALANCE - Moneyhub balance for the account
```

## [Nutmeg](nutmeg)

Connect to [Nutmeg](https://authentication.nutmeg.com/login), login, process one-time-password and return account balance.

```
Retrive Nutmeg balance via web scraping

Usage:
  bin/nutmeg [options]

Options:
  -headless
    	Headless mode (default true)
  -otpcleancommand string
    	Command to clean previous one time password
  -otpcommand string
    	Command to get one time password
  -otppath string
    	Path to file containing one time password message (default "otp/nutmeg")
  -password string
    	Nutmeg password
  -username string
    	Nutmeg username

Environment variables:
  $HEADLESS - Headless mode
  $OTP_CLEANCOMMAND - Command to clean previous one time password
  $OTP_COMMAND - Command to get one time password
  $OTP_PATH - Path to file containing one time password message
  $NUTMEG_USERNAME - Nutmeg username
  $NUTMEG_PASSWORD - Nutmeg password
```

## [Octopus Wheel](octopuswheel)

Connect to [Octopus Energy](https://octopus.energy/login/), login and spin wheel of furtune.

```
Spin octopus wheel of fortune via web scraping

Usage:
  bin/octopuswheel [options]

Options:
  -headless
    	Headless mode (default true)
  -password string
    	Octopus password
  -username string
    	Octopus username

Environment variables:
  $HEADLESS - Headless mode
  $OCTOPUS_USERNAME - Octopus username
  $OCTOPUS_PASSWORD - Octopus password
```

# Desktop Testing

The Makefile contains some test rules which can be enabled by creating an `env` file.  Since this will contain usernames and passwords
my version is not checked in.

```
HEADLESS=false

HA_USERNAME=xxx
HA_PASSWORD=xxx

TEST1_HA_URL=http://homeassistant.local:8123/dashboard-screenshots/0
TEST1_HA_CSS=div.card:nth-child(2) > hui-card:nth-child(1) > plotly-graph:nth-child(1)

TEST2_HA_URL=http://homeassistant.local:8123/dashboard-screenshots/0
TEST2_HA_CSS=div.card:nth-child(6) > hui-card:nth-child(1) > plotly-graph:nth-child(1)

TEST3_HA_URL=http://homeassistant.local:8123/lovelace/overview
TEST3_HA_CSS=div.card:nth-child(5) > hui-card:nth-child(1) > plotly-graph:nth-child(1)

AVIVA_USERNAME=xxx
AVIVA_PASSWORD=xxx
AVIVA_OTPCLEANCOMMAND=ssh xxx rm src/webscrapers/otp/aviva
AVIVA_OTPCOMMAND=scp xxx:src/webscrapers/otp/aviva otp/aviva

AVIVAMYMONEY_USERNAME=xxx
AVIVAMYMONEY_PASSWORD=xxx
AVIVAMYMONEY_WORD=xxx

TEST1_NUTMEG_USERNAME=xxx
TEST1_NUTMEG_PASSWORD=xxx
TEST2_NUTMEG_USERNAME=xxx
TEST2_NUTMEG_PASSWORD=xxx
NUTMEG_OTPCLEANCOMMAND=ssh xxx rm src/webscrapers/otp/nutmeg
NUTMEG_OTPCOMMAND=scp xxx:src/webscrapers/otp/nutmeg otp/nutmeg

TEST1_FUND=0P0001JLD9
TEST2_FUND=0P0001JLD7

TEST1_MONEYFARM_USERNAME=xxx
TEST1_MONEYFARM_PASSWORD=xxx
TEST2_MONEYFARM_USERNAME=xxx
TEST2_MONEYFARM_PASSWORD=xxx
MONEYFARM_OTPCLEANCOMMAND=ssh xxx rm src/webscrapers/otp/moneyfarm
MONEYFARM_OTPCOMMAND=scp xxx:src/webscrapers/otp/moneyfarm otp/moneyfarm

TEST1_MONEYHUB_USERNAME=xxx
TEST1_MONEYHUB_PASSWORD=xxx
TEST1_MONEYHUB_ACCOUNT=My SIPP [ Manual ]

TEST1_OCTOPUS_USERNAME=xxx
TEST1_OCTOPUS_PASSWORD=xxx
```

# Pension scripting

I use these tools to download my pension values, update home assistant, write to google drive (to be picked up by a google spreadsheet) and
send a signal message with the summary.  You'll need to figure out a way to (safely) retrieve one time passwords. 