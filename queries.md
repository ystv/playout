# Potentially useful quieries

For our schedule, we will want to ensure that it isn't null for any linear channels. From my understanding this can be reduced to a "gaps and islands problem". Islands being the scheduled playouts and the gaps being the dead-air.

These queries should work with the [test-data](test-data.sql).

### List the schedule including the playout before's sched_end

```
SELECT
	ROW_NUMBER() OVER(ORDER BY scheduled_start, scheduled_end) AS RN,
	scheduled_start,
	scheduled_end,
	LAG(scheduled_end, 1) OVER (ORDER BY scheduled_start, scheduled_end) AS prev_sched_item_end
FROM
	playout.schedule_playouts
WHERE channel_id = 1;
```

### List the schedule, showing when new islands start and giving each island an ID, includes feature in above SQL

```
SELECT
	*,
	CASE WHEN groups.prev_item_sched_end >= scheduled_start THEN false ELSE true END AS island_start_indicator,
	SUM(CASE WHEN groups.prev_item_sched_end >= scheduled_start THEN 0 ELSE 1 END) OVER (ORDER BY groups.RN) AS island_id
FROM
(
	SELECT
		ROW_NUMBER() OVER(ORDER BY scheduled_start, scheduled_end) AS RN,
		scheduled_start,
		scheduled_end,
		LAG(scheduled_end, 1) OVER (ORDER BY scheduled_start, scheduled_end) AS prev_item_sched_end
	FROM
		playout.schedule_playouts
	WHERE channel_id = 1
) groups
```

### Group islands, creates array of playouts and their start / end

```
SELECT
	array_agg(playout_id) AS playout_ids,
	MIN(scheduled_start) AS island_start,
	MAX(scheduled_end) AS island_end
FROM
(
	SELECT
		*,
		CASE WHEN groups.prev_item_sched_end >= scheduled_start THEN false ELSE true END AS island_start_indicator,
		SUM(CASE WHEN groups.prev_item_sched_end >= scheduled_start THEN 0 ELSE 1 END) OVER (ORDER BY groups.RN) AS island_id
	FROM
	(
		SELECT
			ROW_NUMBER() OVER(ORDER BY scheduled_start, scheduled_end) AS RN,
			playout_id,
			scheduled_start,
			scheduled_end,
			LAG(scheduled_end, 1) OVER (ORDER BY scheduled_start, scheduled_end) AS prev_item_sched_end,
			LAG(playout_id, 1) OVER (ORDER BY scheduled_start, scheduled_end) AS prev_playout_id
		FROM
			playout.schedule_playouts
		WHERE channel_id = 1
	) groups
) islands
GROUP BY
	island_id
ORDER BY
	island_start;
```

### List gaps

```
SELECT
	prev_playout_id,
	playout_id,
	prev_item_sched_end,
	scheduled_start
FROM
```