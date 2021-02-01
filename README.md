# playout

A potentional multi-pipeline scheduled playout system.

## Notes

* [Query development](queries.md)
* Uses the system's local time for scheduler and querying

## Dev notes

* Are we wanting to separate inverse the channel scheduler relationship?  
This would be where a channel would store a schedule_id instead of a schedule storing a channel_id.  
This then introduced a relationship where a channel will
always have a schedule, which isn't always true (could just be nullable on the database but I'd prefer if channel was the root element).  
See the issue is that they both don't need to know about each other or well they have equal importance making it hard to decide which one is the top dog.


## Dependencies

* [postgres](https://www.postgresql.org/)
* [vt](https://github.com/ystv/video-transcode)
* [brave](https://github.com/bbc/brave)

## Building

Developed from Go 1.13+

`go build ./cmd/playout`  
`./playout`

## Simplified overview

The repo offers a bukly application `cmd/playout` which produces an MCR. Possibly in the future each module could be build separately.

MCR manages and groups channels

Channels are the video pipes which ingest a source then create multiple renditions (future versions will have backup source support and slate card support). It doesn't introduce much overhead.

Channels also have two extra optional modules:
* A Piper
* A Scheduler

The piper acts as a safety buffer to a channel's ingest, and swapping between different sources but keeping the output always on.

The scheduler will provide a television schedule to a channel so it will have content to play that out.
* A subroutine which will trigger piper to swap sources to what is on the schedule
* Triggers a player to the channel's ingest (which can be proxied by piper).

Player will playout a programme.

## Design

> Also don't forget to checkout the schema as well. It is heavily documented and commented.

> This might be out-of-date, so check code / schema or simplified overview.

Made of a few elements
### Public-API
* Provides endpoints which shows a public schedule of what is coming

### Manager
* Provides a manual way to write programme playbooks.
* Some management functions to control channels on the fly

### Programme playbooks
* Playlist of video URLs
* A playlist of idents as well
* Time accurate, with time starting at 0, not actual time in-case something has caused a delay.
* Occupies 1 event in the schedule
* It is given to the playout generator, nothing else
    * Playout generator doesn't actually need to look at it techncally. But it's probably worth while figuring what to play.

### Auto-scheduler
* When we don't have content, generate programme playbooks or re-use existing

### Piper
* This will handle feeding the ingest of channel
* Follows the schedule exactly
* Is a safety buffer
* Vision mixing / cutting
* So it will pipe from what the user has set as the input of going out either an rtmp pull or a static video file, it doesn't care. Then act as a safety buffer creating an output on the ingest feed url
* If it's source dies it will cut to the appointed slate video of a channel and alert the channel we've lost the source.

### Channel
* This will take an ingest feed and produce a lot of transcoding jobs in order to generate a DASH output, and optional archival / VCR.
* It is the public face of the actual video output.
    * There are some generic characteristics
    * It will inherit it's properties from the schedule though

### Player (playout generator)
* This is an automatic element
    * So it could be replaced by something manual if necessary?
* Will receive a programme playbook and attempt to follow it and output a signal to an ingest point (not the channel ingest point) where piper will pick it up and play it out.
* The actual software could be something like ffmpeg, obs, ffplayout, liquidsoap. It doesn't really matter. This client will need to just POST information of when it starts and finishes an event.
