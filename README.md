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
