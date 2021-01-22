-- Create two generic channels of the two types
INSERT INTO playout.channel
(name, description, type, ingest_url, slate_url,
playback_url, visibility, archive, dvr)
VALUES
('The TV Channel', 'Top-end content', 'linear',
'rtp://media-ingest.ystv.co.uk/sky1',
'https://cdn.ystv.co.uk/channel-assets/holding.mp4',
'https://media-serve.ystv.co.uk/sky1/manifest.m3u8',
'public', true, true),
('MusicTV', 'A pop-up music channel!', 'event',
'rtmp://media-ingest.ystv.co.uk/mtv',
'https://cdn.ystv.co.uk/channel-assets/holding.mp4',
'https://media-serve.ystv.co.uk/mtv/manifest.m3u8',
'public', true, false);
--
-- Generate some programmes
INSERT INTO playout.programmes
(title, description, thumbnail)
VALUES
('Cooking Time!', 'We are in a Kitchen, so sandwiches',
'https://cdn.ystv.co.uk/prog-assets/101.jpg'),
('Bird Documentary', 'Big pigeons',
'https://cdn.ystv.co.uk/prog-assets/103.jpg'),
('Heavy Pop', 'Like heavy rock but pop',
'https://cdn.ystv.co.uk/prog-assets/220.jpg'),
('Top Crimbo', 'The best christmas hits',
'https://cdn.ystv.co.uk/prog-assets/560.jpg'),
('Funny comedy', 'epic funny comedy',
'https://cdn.ystv.co.uk/prog-assets/348.jpg');
--
-- Make a little schedule
INSERT INTO playout.schedule
(channel_id, programme_id, ingest_url,
scheduled_start, scheduled_end, name)
VALUES
(1, 2, 'rtp://media-land.ystv.co.uk/vod/c_s01_e01.mp4',
'2020-01-21 09:00:00.000', '2020-01-21 09:45:00.000',
'Documentary: The pigeons'),
(1, 5, 'rtp://media-land.ystv.co.uk/vod/com_s43_e02.mp4',
'2020-01-21 09:45:00.000', '2020-01-21 10:00:00.000',
'Comedy: Something very very funny!'),
(1, 1, 'rtp://media-ingest.ystv.co.uk/ob2',
'2020-01-21 10:30:00.000', '2020-01-21 12:00:00.000',
'Cooking Time with Rhys!'),
(2, 3, 'rtp://media-ingest.ystv.co.uk/ob1',
'2020-01-21 09:00:00.000', '2020-01-21 11:00:00.000',
'Hard Pop?'),
(2, 4, 'rtp://media-land.ystv.co.uk/vod/crimbo.mp4',
'2020-01-21 11:00:00.000', '2020-01-21 12:00:00.000',
'Crimbo time');