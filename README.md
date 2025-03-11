# veloherodown



Create a local copy of your [Velo Hero](https://www.velohero.com/) data.

![Velo Hero Logo](https://www.velohero.com/static/touchicon.png)

This Go application creates an export of your recorded activities at Velo Hero.
The first time all the files are downloaded.
For further calls only changes and new files are downloaded.
The export is stored in your chosen format(s): `JSON`, `PWX`, `CSV`, `GPX`, `KML`, or `TCX`.
The JSON file contains all the details except the comments of other users.
The PWX file also has many details and can be processed by [Golden Cheetah](http://www.goldencheetah.org/).
The filename is the ID of the activity (`https://app.velohero.com/workouts/show/<ID>`).

## Setup

1. Create a directory for your Velo Hero export
1. Download the veloherodown application for your operating system and architecture
1. Rename it to:
    * `velohero` (macOS, Linux)
    * `velohero.exe` (Windows)
1. Go to <https://app.velohero.com/sso> to get your private single sign-on key
1. Run the application - it will prompt you to enter your SSO key
    * Alternatively, create a `.veloherorc` file in the directory with:
        ```ini
        VELOHERO_SSO_KEY=[insert your own]
        ```

## Usage

Start export:

```bash
veloherodown [FORMAT]
```

Replace `[FORMAT]` with one or a set of

* `json`: Velo Hero generic JSON format (with all details)
* `pwx` : Training Peaks PWX file with laps (can be processed by Golden Cheetah)
* `csv` : Comma-Seperated Values CSV file
* `gpx` : GPX file (only the geo coordinates)
* `kml` : Google Earth KML file
* `tcx` : Garmin TCX file

The default format is PWX.

Example:

```bash
veloherodown json pwx
```

All files will be downloaded to the current directory where the `.veloherorc` file is located.

## Features

* Downloads only new or changed activities since the last run
* Supports multiple export formats
* Automatically skips already downloaded files
* Respects server load with appropriate delays between requests
* Simple configuration with a single SSO key

## Notes

The application stores a tracking file `.velohero_last_export.do_not_remove` in the current directory to keep track of the last export timestamp. Do not delete this file if you want incremental updates.