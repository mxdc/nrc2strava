# NRC2Strava

This project exports run activities from **Nike Run Club (NRC)** and imports them into **Strava**. It supports both outdoor and indoor run activities.

Initially, I used [nrc-exporter](https://github.com/yasoob/nrc-exporter) to import my NRC runs. However, treadmill runs were not supported due to the limitations of the `.gpx` format. To address this, I created this project to convert NRC runs into the **FIT format**, which is more advanced and fully compatible with Strava.

Features:
- Downloads Nike Run Club activities and saves them to disk.
- Converts these activities from the **.json** format to the **.fit** format, compatible with the Strava platform.
- Handles both indoor (treadmill) and outdoor (GPS) runs.
- Uploads **.fit** activities to Strava without requiring you to create an app on the developer platform.

## Requirements

- **Go**: Ensure you have Go installed on your system.

## Step by Step

We will import the entire NRC activity history into Strava by following these steps.

### 0. Build

Clone and compile:
```bash
$ git clone git@github.com:mxdc/nrc2strava.git
$ cd nrc2strava
$ go mod tidy
$ go build -o bin/nrc2strava cmd/main.go
```

### 1. Download NRC Activities

**Retrieve the NRC Token**

To download activities from NRC, you need to obtain your **Nike access token**.

First, log in to [Nike.com](https://www.nike.com/) with your account using your web browser.

You have two options:
* Navigate to the **Application** tab, locate `oidc.user:https://accounts.nike.com:4fd2d5e7db76e0f85a6bb56721bd51df` in Local Storage, and copy the `access_token`.
* Retrieve the token from the browser's developer console using the following command:
```javascript
JSON.parse(window.localStorage.getItem('oidc.user:https://accounts.nike.com:4fd2d5e7db76e0f85a6bb56721bd51df')).access_token
```

Once you have the token, export it as an environment variable:
```bash
$ export NIKE_TOKEN='<access_token>'
```

**Download NRC activities**

Once you have the token, use the following command to download activities:
```bash
$ bin/nrc2strava download --activities.dir='./downloaded' --nrc.token="$NIKE_TOKEN"
```

This will save all activities as JSON files in the `./downloaded` directory.

**Convert JSON Activities to FIT Format**

With all your activities now saved on disk as JSON files, you can convert them into FIT files:
```bash
$ bin/nrc2strava convert --activities.dir './downloaded' --fit.dir './output'
```

The FIT files will be saved in the `./output` directory.

### 2. Upload FIT Activities to Strava

**Retrieve the Strava Tokens**

The Strava tokens are stored in the browser cookies.

Log in to [Strava](https://www.strava.com/), open the developer console, navigate to the **Application** tab, and locate the `_strava4_session` cookie.

Export this value as an environment variable:
```bash
$ export STRAVA4SESSION='<_strava4_session>'
```

**Upload FIT Activities to Strava**

Upload the FIT activities to Strava:
```bash
$ bin/nrc2strava upload --fit.dir='./output' --strava.token="$STRAVA4SESSION"
```

Each successfully uploaded `.fit` file is automatically moved to an `uploaded` subfolder. This allows you to run the command multiple times safely without re-uploading the same files.

> **Note:** If you have more than 600 run activities, the Strava API may rate limit requests and return HTTP 429.

### 3. Download Activities from Strava

Optionally, download all your activities from Strava to your local disk:
```bash
$ bin/nrc2strava strava-download --activities.dir='./strava-downloaded' --strava.token="$STRAVA4SESSION"
```
