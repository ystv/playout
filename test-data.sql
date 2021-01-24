-- Create two generic channels of the two types
INSERT INTO playout.channel
(name, description, type, ingest_url, ingest_type, 
slate_url, visibility, archive, dvr)
VALUES
('The TV Channel', 'Top-end content', 'linear',
'rtp://media-ingest.ystv.co.uk/sky1', 'rtp',
'https://cdn.ystv.co.uk/channel-assets/holding.mp4',
'public', true, true),
('MusicTV', 'A pop-up music channel!', 'event',
'rtmp://media-ingest.ystv.co.uk/mtv', 'rtmp',
'https://cdn.ystv.co.uk/channel-assets/holding.mp4',
'public', true, false);
--
-- Generate some programmes
INSERT INTO playout.programmes
(title, description, thumbnail, type)
VALUES
('Cooking Time!', 'We are in a Kitchen, so sandwiches',
'https://cdn.ystv.co.uk/prog-assets/101.jpg', 'live'),
('Bird Documentary', 'Big pigeons',
'https://cdn.ystv.co.uk/prog-assets/103.jpg', 'vod'),
('Heavy Pop', 'Like heavy rock but pop', 
'https://cdn.ystv.co.uk/prog-assets/220.jpg', 'live'),
('Top Crimbo', 'The best christmas hits',
'https://cdn.ystv.co.uk/prog-assets/560.jpg', 'vod'),
('Funny comedy', 'epic funny comedy',
'https://cdn.ystv.co.uk/prog-assets/348.jpg', 'vod');
--
-- Make a little schedule
INSERT INTO playout.schedule_blocks
(channel_id, programme_id, ingest_url, ingest_type,
scheduled_start, scheduled_end)
VALUES
(1, 2, 'rtp://media-land.ystv.co.uk/player/e45t', 'rtp',
'2020-01-21 09:00:00.000', '2020-01-21 09:45:00.000'),
(1, 5, 'rtp://media-land.ystv.co.uk/player/ftge', 'rtp',
'2020-01-21 09:45:00.000', '2020-01-21 10:00:00.000'),
(1, 1, 'rtp://media-ingest.ystv.co.uk/ob2', 'rtp',
'2020-01-21 10:30:00.000', '2020-01-21 12:00:00.000'),
(2, 3, 'rtp://media-ingest.ystv.co.uk/ob1', 'rtp',
'2020-01-21 09:00:00.000', '2020-01-21 11:00:00.000'),
(2, 4, 'rtp://media-land.ystv.co.uk/player/j67u', 'rtp',
'2020-01-21 11:00:00.000', '2020-01-21 12:00:00.000');