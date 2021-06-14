# Crispy Musicular

Utility to backup Spotify and Youtube playlists locally.

Due to a growing amount of songs being stored in Spotify playlists the event of losing
those playlists would be pretty devastating. This tool is aimed to alleviate that risk by
making regular backups of all user playlists and storing them somewhere safe so if such
thing ever happens, the actual song names are still preserved and playlists can be recreated
either on other platform or just by acquiring songs.

Since not all songs exist on Spotify, Youtube is a great resource for finding a lot of the rare
tracks. But at the same time videos on Youtube disappear a lot more often. This tool is now also
capable of preserving specified Youtube playlists.

![App](/sc.jpg "Screenshot")

## About

Application has 2 main components, HTTP frontend and backup package.

### Frontend

Frontend is responsible for authenticating user and has some utility features
for updating config, starting manual backup and showing basic config stats.

#### Spotify Authentication

Only one user can be authenticated with the application
and the application runs backups only for that user. User is authenticated using
oauth and dev app created using spotify dev program. Using this method only a single
authentication is enough to run the application indefinitely, unless refresh token is
somehow revoked.

Client ID and Secret can be obtained at Spotify dev page: https://developer.spotify.com/dashboard/applications

```
SPOTIFY_ID=spotify_app_id
SPOTIFY_SECRET=spotify_app_secret
# inside config.yaml
spotifyCallback: http://localhost:3333/callback
```

#### Youtube Authentication

Only one user can be authenticated with the application
and the application runs backups only for that user. User is authenticated using
oauth and dev app created using google dev. Using this method only a single
authentication is enough to run the application indefinitely, unless refresh token is
somehow revoked.

These can be obtained by [creating a project](https://developers.google.com/workspace/guides/create-project) and [obtaining desktop application credentials](https://developers.google.com/workspace/guides/create-credentials#desktop).

Youtube Data specific part [can be found here](https://developers.google.com/youtube/v3/getting-started).

Once configured provide env variables and update config file:

```
YOUTUBE_ID=youtube_app_id
YOUTUBE_SECRET=youtube_app_secret
# inside config.yaml
youtubeCallback: http://localhost:3333/youtube/callback
```

### Backup package

Utilizes go channels to make it concurrent
Using go channels X amount of workers is created which then received all the users playlists.
Based on config options it checks which playlists should be backed up and then each worker works
on a single playlist at a time.

Once all playlists are saved, post backup actions are run. If during saving there are any errors, a backup is deemed invalid and post backup actions are not
run.

Performance on my machine is not bad, running a backup with 8 workers on 41 playlists with total of
4.2k tracks takes ~3-5seconds. This might be impacted by API ratelimit being breached and other factors,
but it is definitely good enough. With 1 worker it ran for about 12 seconds.

#### Post backup actions

There are currently 2 backup actions:

##### JSON Backup action

Enabled by providing in the config values:

```yaml
# Whether it is enabled
jsonActionEnabled: false
# Where to store json files
jsonDir: json/
```

If enabled, this will serialize backup as a json file and store it in the provided directory.


##### Google Drive backup action

Enabled by providing in the config values:

```yaml
# Whether it is enabled
driveActionEnabled: true
# Callback url for oauth2 flow, shouldn't change unless application changes
driveCallback: http://localhost:3333/drive/callback
# Optional, default "crispy_spotify_backups"
driveDir: directory_name
```

Additionally env variables have to be provided:

```sh
DRIVE_ID=google_drive_app_id
DRIVE_SECRET=google_drive_app_secret
```

These can be obtained by [creating a project](https://developers.google.com/workspace/guides/create-project) and [obtaining desktop application credentials](https://developers.google.com/workspace/guides/create-credentials#desktop).

Scope for the application/credentials should be `https://www.googleapis.com/auth/drive.file`. This scope only allows application to touch files that it created or which were shared with it, so technically theres no chance for it to touch and/or ruin any other files.

If enabled, this will create a directory with a provided `driveDir` name and keep writing JSON style backups there after each backup.

### Logging

Logs are written to STDOUT and also a file.
Log file directory can be configured with env variable `LOG_DIR`, where log file will be at `LOG_DIR/crispy.log`.
Currently there is not log file rollover or truncation, it will only be appended.

### Known errors

#### Spotify Auth

Due to some strange reason Spotify oauth endpoint sometimes returns 503 error when trying to refresh token.
This is intermittent and probably can only be resolved by spotify or by retrying certain actions manually.

`oauth2: cannot fetch token: 503 Service Unavailable`

#### Permissions

Some permission realated issues can arise.
For example SQLITE needs proper RW access to folder and files `[dbname]-wal`, `[dbname]-shm` for it to work,
otherwise it might error out that `"database is in read-only mode"`.

Same thing also applies to JSON folder, application needs to have write access there as well.

Though generally this should not be an issue.

## Project layout

Project layout based on: https://github.com/golang-standards/project-layout
Also: https://github.com/katzien/go-structure-examples/

```
build/ - docker related things for building image
pkg/ - shared code
cmd/ - final binary that runs the code
templates/ - templates used for http frontend
```

## Config

Configuration is done using .yaml file. Code for it exists at `pkg/config/config.go`

```yaml
# How long to wait between backup runs
runIntervalSeconds: 1800
# Port on which HTTP server will listen
port: 3333
# Spotify callback to be used for auth
spotifyCallback: http://localhost:3333/callback
# Backup concurrent worker count
workerCount: 8
# How long to wait for workers to finish
workerTimeoutSeconds: 600
# Playlists to save
savedPlaylistIds: []
# Playlists to ignore
ignoredPlaylistIds: []
# Whether to ignore playlists not created by user itself
ignoreNotOwnedPlaylists: true
# json backup output directory
jsonDir: json/
# path to datbase file
dbPath: data/a.db
### Youtube Settings
youtubeSavedPlaylistIds:
    - LL # For Liked videos
youtubeCallback: http://localhost:3333/youtube/callback
### Google Drive Settings
driveActionEnabled: true
driveCallback: http://localhost:3333/drive/callback
driveDir: crispy_spotify_backups
```

## Backup storage

Backups are stored in 2 places, SQLite database and as JSON.

### Database

Main tables used/created in the SQLite database:
- `backups` - stores general entry about the backup
- `playlists` - stores entries for each playlist and relation to backup
- `tracks` - stores entries for each track and relation to playlist and backup
- `youtube_playlists` - same as above, but for youtube
- `youtube_tracks` - same as above, but for youtube

Other tables:
- `auth_state` - stores persisted state about authenticated user so that after service reboot user would not need to re-authenticate.

Example query to get tracks of certain backup. This can be used to create a list of spotify URI's to quickly re-create a playlist.
```
sqlite> SELECT p.name, t.artist, t.name, t.album FROM tracks t JOIN playlists p ON p.id = t.playlist_id WHERE t.backup_id = 1;
groovy soul/funk|Patrice Rushen|Remind Me|Straight From The Heart
groovy soul/funk|Patrice Rushen|Settle For My Love|Pizzazz
groovy soul/funk|The Jones Girls|When I'm Gone|At Peace with Woman
groovy soul/funk|The Jones Girls|Who Can I Run To|The Jones Girls
groovy soul/funk|Dexter Wansel|The Sweetest Pain|Time Is Slipping Away
groovy soul/funk|Keni Burke|Risin' to the Top|Changes (Expanded Edition)
groovy soul/funk|Rene & Angela|I Love You More - Remastered|Classic Masters
groovy soul/funk|Evelyn "Champagne" King|The Show Is Over|Smooth Talk (Expanded Edition)
```

### JSON Output

Using `Playlists` array and `Tracks` array which contains objects with property `PlaylistId` it is trivial to
correlate which tracks belong to which playlist.

```
{
  "Backup":{
    "Id":1,
    "UserId":"...",
    "Success":true,
    "Started":"2021-05-03T09:36:34.451335776Z",
    "Finished":"2021-05-03T09:36:37.523334611Z"
  },
  "Playlists":[
    {
      "Id":1,
      "SpotifyId":"...",
      "Name":"...",
      "Created":"2021-05-03T09:36:34.728931355Z"
    },
    ...
  ],
  "Tracks":[
    {
      "Id":1,
      "SpotifyId":"3vc0dm7NHZTProvlYlkhmh",
      "Name":"Journal of Ardency",
      "Artist":"Class Actress",
      "Album":"Journal of Ardency",
      "AddedAtToPlaylist":"2021-03-18T07:56:39Z",
      "Created":"2021-05-03T09:36:34.729091604Z",
      "PlaylistId":1
    },
    {
      "Id":2,
      "SpotifyId":"1RbCFHtxDRmaFR7HAUMGtp",
      "Name":"Weekend",
      "Artist":"Class Actress",
      "Album":"Rapprocher",
      "AddedAtToPlaylist":"2021-03-18T07:57:23Z",
      "Created":"2021-05-03T09:36:34.729172414Z",
      "PlaylistId":1
    },
    ...
  ]
}
```

## Docker

App can be easily built and ran as docker image.
Besides basic configuration it is also required to setup following spotify env vars
which can be obtainted from spotify dev page: https://developer.spotify.com/dashboard/applications

```
SPOTIFY_ID=spotify_app_id
SPOTIFY_SECRET=spotify_app_secret
DRIVE_ID=google_drive_app_id
DRIVE_SECRET=google_drive_app_secret
YOUTUBE_ID=youtube_app_id
YOUTUBE_SECRET=youtube_app_secret
```

basic steps to do that are as follows:
```sh
#!/bin/sh
IMAGE_NAME=crispy

docker build -f build/package/Dockerfile . -t "$IMAGE_NAME"

# each line of docker run explained:
# mount where json will be saved
# mount where config exists (should be a directory, direct file mount doesn't work for updating config)
# mount where sqlite db exists
# config path
# load other ENV vars from file, or use -e X=Y
# expose port (depends on config)
# run built image
docker run --rm -it \
  -v "$PWD/json":"/go/src/app/json" \
  -v "$PWD/conf":"/go/src/app/conf" \
  -v "$PWD/data":"/go/src/app/data" \
  -e CONFIG_PATH="/go/src/app/conf/conf.yaml" \
  --env-file=".env.local" \
  -p 3333:3333 \
  "$IMAGE_NAME"
```
