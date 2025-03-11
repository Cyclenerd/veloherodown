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

1. Sign up at <https://app.velohero.com/sso>
2. Get yourself a private single sign-on key. That's the long string.
3. Create a `.veloherorc` file in the directory where you want to store your exports. Save the SSO key in this file::

```ini
VELOHERO_SSO_KEY=[insert your own]
```

## Usage

Start export:

```bash
veloherodown [FORMAT]
```

Replace `[FORMAT]` with one or a set of

* `json`: Velo Hero generic format
* `pwx` : Trainings Peaks PWX
* `csv` : Comma-Seperated Values
* `gpx` : GPX track
* `kml` : Google Earth KML
* `tcx` : Garmin TCX

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