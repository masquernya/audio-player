# audio player

Basic linux audio player that values launch speed and displaying a visualization of the audio to allow scrubbing through audio files.

## comparison to other programs

- **Audacity**
  - Audacity is slower to launch
  - It's also kinda buggy/unreliable (Ubuntu deb sometimes doesn't launch for me, [AppImage doesn't work](https://github.com/audacity/audacity/issues/6233))
- **VLC**
  - Somewhat buggy for me, and there is no support for visualizations/scrubbing.
- **Every Other Linux Audio Player**
  - They all seem to value playlists for some reason (e.g. Sayonara auto loops if there's only one song being played (and I can't seem to turn that off), Elisa resumes the previous song rather than playing the selected file, Rhythmbox scans your disk and insists on playing a new song after playing a file)
- **CLI** (e.g. `ffplay`)
  - Faster/more reliable, but doesn't have a UI or visualization.

## runtime dependencies

- `ffmpeg`, `ffprobe`, and `ffplay` (ffplay and ffprobe are usually included with ffmpeg)