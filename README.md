# NRC2Strava

This project exports run activities from **Nike Run Club (NRC)** and imports them into the **Strava platform**. It supports both outdoor and indoor run activities.

Initially, I used the project [nrc-exporter](https://github.com/yasoob/nrc-exporter) to import my NRC runs. However, treadmill runs were not supported due to the limitations of the `.gpx` format. To address this, I started this project to convert NRC runs into the **FIT format**, which is more advanced and fully compatible with Strava

Specific to this project:
- Uses the **.FIT file format**
- Compatible with both indoor (treadmill) and outdoor run activities.
- Works with regular tokens, no need to create a Strava App.

The process is divided into 3 steps:

1. **Download**: Import the run activities from NRC and save each one as a JSON file on the local disk.
2. **Convert**: Convert the JSON activities into the FIT format (compatible with Strava) and save them on the local disk.
3. **Upload**: Upload the FIT activities from the local disk to Strava, one by one.

To execute all the steps in a single operation:

1. **Migrate**: Perform all the above steps in a single command.

## Requirements

- **Go**: Ensure you have Go installed on your system.
- **Nike Access Token**: Required to download activities from NRC.
- **Strava Session Tokens**: Required to upload activities to Strava.


## Credentials

**Retrieve the NRC Token**

To download activities from NRC, you need to obtain your **Nike access token**.
You have 2 options:

Option A: Log in to nike.com with your account, navigate to the **Application** tab, locate the `oidc.user:https://accounts.nike.com:4fd2d5e7db76e0f85a6bb56721bd51df` in the Local Storage, and copy the `access_token`.

Option B: Log in to nike.com with your account, then retrieve the token from the browser's developer console using the following command:
```javascript
JSON.parse(window.localStorage.getItem('oidc.user:https://accounts.nike.com:4fd2d5e7db76e0f85a6bb56721bd51df')).access_token
```

Once you have the token, export it as an environment variable:
```bash
$ export NIKE_TOKEN='<access_token>'
```

**Retrieve the Strava Tokens**

The Strava tokens are stored in the browser cookies.
Open the developer console, navigate to the **Application** tab, and locate the `_strava4_session` and `xp_session_identifier` cookies.

Export these values as environment variables:
```bash
$ export STRAVA4SESSION='<_strava4_session>'
$ export XPSESSIONIDENTIFIER='<xp_session_identifier>'
```

## Build

Compile the source code:
```bash
$ go build -o bin/nrc2strava cmd/main.go
```


## Steps

### 1. Download Activities from NRC

Once you have the token, use the following command to download activities:
```bash
$ go run ./cmd/main.go download --activities.dir='./downloaded' --nrc.token="$NIKE_TOKEN"
```

This will save all activities as JSON files in the `./downloaded` directory.


### 2. Convert JSON Activities to FIT Format

Convert a single JSON activity to a FIT file:
```bash
$ go run ./cmd/main.go convert --activity.file './downloaded/run-01.json'
$ go run ./cmd/main.go convert --activity.file './downloaded/run-02.json'
```

Convert multiple JSON activities to FIT files:
```bash
$ go run ./cmd/main.go convert --activities.dir './downloaded' --fit.dir './output'
```

The FIT files will be saved in the `./output` directory.


### 3. Upload FIT Activities to Strava

Upload a single FIT activity to Strava:
```bash
$ go run ./cmd/main.go upload --fit.file='./output/run-01.fit' --strava.token="$STRAVA4SESSION" --strava.id="$XPSESSIONIDENTIFIER"
```

Upload multiple FIT activities to Strava:
```bash
$ go run ./cmd/main.go upload --fit.dir './output' --strava.token="$STRAVA4SESSION" --strava.id="$XPSESSIONIDENTIFIER"
```


### 4. Migrate Activities from NRC to Strava

To perform all the steps (download, convert, and upload) in one command:
```bash
$ go run ./cmd/main.go migrate \
    --fit.dir './tmp' \
    --nrc.token="$NIKE_TOKEN" \
    --strava.token="$STRAVA4SESSION" \
    --strava.id="$XPSESSIONIDENTIFIER"
```
