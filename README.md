# Crispy Musicular

Utility to backup Spotify playlists locally.

Due to a growing amount of songs being stored in Spotify playlists the event of losing
those playlists would be pretty devastating. This tool is aimed to aleviate that risk by
making regular backups of all user playlists and storing them somewhere safe so if such
thing ever happens, the actual song names are still preserved and playlists can be recreated
either on other platform or just by acquiring songs.

## Ideas:

Retrieve user playlists, iterate through them, save playlist metadata + each song in the playlist.
Either use JSON or maybe SQLite?
Make it possible to automatically sync to some more reliable storage (GDrive?)?
Do incremental diff's instead off full copies?

Make it "self contained" with one time setup required (token refresh/expiration?).
In a sense that it starts then waits for user to visit the site and complete oauth flow. Once that
is done the user "owns" the application and any other subsequent authentications do not impact for
whom the backup is being done. To stop the backuping either destroy the storage that is being used for
tracking backup work (SQLITE DB?) or maybe add a button for the owner to disconnect/stop.

## More

On startup:

- Load state from X
- IF no state exists - allow linking account
- IF state exists - disallow linking, allow unlinking

- Load configuration (Spotify ClientId/Secret, Backup interval, backups to keep?)
- Create a timer/ticker for interval
- Once ticker hits, start a go routine for backing up

  - Backup Routine
  - Create In channel for getting playlist Id/basic info
  - Query UserPlaylists
  - Send user playlist names to workers (through channel) that query playlists


## Notes

Project layout based on: https://github.com/golang-standards/project-layout
Also: https://github.com/katzien/go-structure-examples/

```
build/ - docker related things for building image
pkg/ - shared code
cmd/ - final binary that runs the code
```

Initial version should just store backed up information, with maybe some general aggregated information returned.

### Domain

- User - Spotify User
- Backup - Instance of backup linked to user, contains:
  - Playlist - Spotify Playlist
  - Song - Spotify Song

Needs backup runner / main loop

### Database

Tables:
- Songs, main table with Id (Sequence), SongId (Spotify), Name, Artists, DateAddedToPlaylist (If Available), PlaylistId (Since user can control these), SpotifyUri (For rebuilding playlist), BackupId
- Playlists, Id (Sequence), PlaylistId (Spotify), Name, Created (Or Smth)
- Backups, Id (Sequence), Started, Finished, (Some stats?)

Some thoughts:
- Songs stores straight up names and artist names as strings, because technically the label can change these so the backup will contain the string of what it was called at that specific time.
- Creating another table for Song that can be related to BackupSong would be possible and technically would save a decent amount of space since there would be no need to constantly save entire song name+artist, just ID's, but that just seems too much effort atm and also would be annoying to deal, as it would require to first insert/check into Song table and then another insert into SongBackup table, also updating Song in case some details change.
