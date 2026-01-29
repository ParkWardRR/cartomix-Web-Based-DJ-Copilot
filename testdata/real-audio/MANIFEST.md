# Real Audio Test Corpus

This directory contains real audio files for testing the Algiers analyzer.
**Do NOT commit these files to git.** They are for local testing only.

## Track Summary

| Track | Artist | Album | Duration | Sample Rate | Bit Depth | Genre |
|-------|--------|-------|----------|-------------|-----------|-------|
| With All The World | Khruangbin | Hasta El Cielo | 00:03:43.503 | 96.0 kHz | 24 bits | Dub |
| Ain't No Sunshine | Bill Withers | Just As I Am | 00:02:05.093 | 44.1 kHz | 16 bits | Soul |
| Grandma's Hands | Bill Withers | Just As I Am | 00:02:01.107 | 44.1 kHz | 16 bits | Soul |
| How I Love | Khruangbin | Hasta El Cielo | 00:04:36.832 | 96.0 kHz | 24 bits | Dub |
| Queen Tings | Masego & Tiffany Gouché | Lady Lady | 00:03:08.053 | 44.1 kHz | 16 bits |  |
| Cold Little Heart | Michael Kiwanuka | Love & Hate | 00:09:57.600 | 96.0 kHz | 24 bits | Soul / Funk / R&B |
| Give Life Back to Music | "Daft Punk, Nathan East, NILE RODGERS, Paul Jackson Jr., Quinn, Greg Liesz, Chris Caswell, CHILLY GONZALES, John ""JR"" Robinson" | Random Access Memories | 00:04:34.408 | 88.2 kHz | 24 bits | Dance |
| Lose Yourself to Dance | "Pharrell Williams, Daft Punk, Nathan East, NILE RODGERS, John ""JR"" Robinson" | Random Access Memories | 00:05:53.868 | 88.2 kHz | 24 bits | Dance |
| Get Lucky | Pharrell Williams, NILE RODGERS, Daft Punk, Nathan East, OMAR HAKIM, Paul Jackson Jr., Chris Caswell | Random Access Memories | 00:06:09.615 | 88.2 kHz | 24 bits | Dance |
| Doin’ It Right | Daft Punk, Panda Bear | Random Access Memories | 00:04:11.317 | 88.2 kHz | 24 bits | Dance |
| Contact | Daft Punk, OMAR HAKIM, James Genus, DJ Falcon | Random Access Memories | 00:06:21.521 | 88.2 kHz | 24 bits | Dance |
| Lady Lady | Masego | Lady Lady | 00:02:34.667 | 44.1 kHz | 16 bits |  |
| Rules - Scientist Dub | Khruangbin | Hasta El Cielo | 00:04:28.008 | 96.0 kHz | 24 bits | Dub |
| Tadow | Masego feat. FKJ | Lady Lady | 00:05:01.893 | 44.1 kHz | 16 bits |  |
| Love & Hate | Michael Kiwanuka | Love & Hate | 00:07:07.093 | 96.0 kHz | 24 bits | Soul / Funk / R&B |

## Source

Tracks copied from: `/Volumes/navidrome-music/Staging/`

## Usage

```bash
# Analyze a single track
.build/release/analyzer-swift analyze testdata/real-audio/*.flac --progress

# Scan entire directory
go run ./cmd/engine scan testdata/real-audio/
```
