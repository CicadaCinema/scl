# SCL - Strava Club Leaderboard

## Setup

Sign up for a Strava account. Go to [https://www.strava.com/settings/api](https://www.strava.com/settings/api) and create an app.

Record the *Client ID* and *Client Secret* for your app

## Configuration

* The environment variables `STRAVA_CLIENT_ID` and `STRAVA_CLIENT_SECRET` are required by scl-api
* Make sure to edit the line `'process.env.STRAVA_CLIENT_ID': "111111",` in `scl-web/rollup.config.js` to include your Strava app's client id instead of `111111`