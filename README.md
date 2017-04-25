# veloherodown

Create a local copy of your [Velo Hero](https://www.velohero.com/) data.

![Velo Hero Logo](https://www.velohero.com/static/touchicon.png)

This Bash script creates an export of your recorded activities at Velo Hero.
The first time all the files are downloaded.
For further calls only changes and new files are downloaded.
The export is stored as a PWX and JSON file.
The filename is the ID of the activity (`https://app.velohero.com/workouts/show/<ID>`).

## Prerequisites

* Bash shell
* curl

Most Linux distributions and macOS meet the requirements.

## Setup

1. Sign up at https://app.velohero.com/sso
2. Get yourself a private single sign-on key. That's the long string.
3. Create a `.veloherorc` file in your home directory. Save the SSO key and the storage location for the export in this file:
~~~
VELOHERO_SSO_KEY=[insert your own]
VELOHERO_EXPORT_DIR=[specify location for export]
~~~

## Usage

Start export:

    veloherodown
