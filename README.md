# playout

A potentional multi-pipeline scheduled playout system.

## Notes

* Uses the system's local time for scheduler and querying

## Dependencies

* postgres
* vt
* brave

## Building

Developed from Go 1.13+

`go build ./cmd/playout`
`./playout`

## Design

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

### Playout generator
* This is an automatic element
    * So it could be replaced by something manual if necessary?
* Will receive a programme playbook and attempt to follow it and output a signal to an ingest point (not the channel ingest point) where piper will pick it up and play it out.
* The actual software could be something like ffmpeg, obs, ffplayout, liquidsoap. It doesn't really matter. This client will need to just POST information of when it starts and finishes an event.