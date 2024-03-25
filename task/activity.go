package task

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/atony2099/pomo/db"
	"github.com/atony2099/pomo/domain"
)

const defaultActivity = "study"
const pomodoroDuration = 20 * time.Minute

// Define the struct to model the daily_trackers table

func GetActivities(offset int) {
	date := time.Now().AddDate(0, 0, -offset)
	day := date.Format("2006-01-02")

	// Fetch activities from daily_trackers for the given date
	var activities, err = db.SelectDailyTracker(day)
	if err != nil {
		fmt.Printf("Error selecting daily tracker from day: %v\n", err)
		return
	}

	fmt.Printf("Activities for %s:\n", day)
	for _, activity := range activities {
		// fmt.Printf("%s: %s - %s\n", activity.Activity, activity.StartTime.Format("15:04:05"), activity.EndTime.Format("15:04:05"))

		// align the output
		fmt.Printf("%-20s: %s - %s\n", activity.Activity, activity.StartTime.Format("15:04:05"), activity.EndTime.Format("15:04:05"))

	}
	// group by activity name

	var maps = make(map[string]time.Duration)
	for _, activity := range activities {
		if _, ok := maps[activity.Activity]; !ok {
			maps[activity.Activity] = activity.EndTime.Sub(activity.StartTime)
		} else {
			maps[activity.Activity] += activity.EndTime.Sub(activity.StartTime)
		}
	}
	fmt.Println("\nTotal duration for each activity:")
	for k, v := range maps {
		// align the output
		fmt.Printf("%-20s: %v\n", k, v)
	}

	fmt.Println()
}

func completeNullEndTime(day string) {
	// Fetch activities from daily_trackers for the given date
	var activities, err = db.SelectDailyTracker(day)
	if err != nil {
		fmt.Printf("Error selecting daily tracker from day: %v\n", err)
		return
	}

	for _, activity := range activities {
		if activity.EndTime == nil {
			// prompt for end time

			reader := bufio.NewReader(os.Stdin)
			fmt.Printf("Enter end time for %s,which start time is %s: ", activity.Activity, activity.StartTime.Format("15:04:05"))
			endTime, _ := reader.ReadString('\n')
			endTime = strings.TrimSpace(endTime)

			if endTime == "" {
				now := time.Now()
				activity.EndTime = &now
			} else {
				hourInt, err := strconv.Atoi(endTime[:2])
				if err != nil {
					log.Fatalf("error converting to int: %v", err)
				}
				minuteInt, err := strconv.Atoi(endTime[2:])
				if err != nil {
					log.Fatalf("error converting to int: %v", err)
				}

				endTimes := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), hourInt, minuteInt, 0, 0, time.Local)
				activity.EndTime = &endTimes

			}
			err = db.UpdateDailyTracker(activity)
			if err != nil {
				log.Fatalf("error updating daily tracker: %v", err)
			}

		}
	}
}
func Complete(offset int) {
	// fmt.Println("Complete task")

	date := time.Now().AddDate(0, 0, -offset)

	// format to 2006-01-02
	day := date.Format("2006-01-02")

	// first complet the activity which end time is null
	completeNullEndTime(day)

	// Fetch activities from time_entries for the given date
	var entries, err = db.SelectTimeEntry(day)
	if err != nil {
		fmt.Printf("Error selecting task from day: %v\n", err)
		return
	}

	// Insert fetched activities into daily_trackers as "study"
	for _, entry := range entries {

		// if entry.start time is privous day, set it to 00:00:00
		if entry.StartTime.Format("2006-01-02") != day {
			entry.StartTime = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		}
		if entry.EndTime.Format("2006-01-02") != day {
			// to last minute of the day
			entry.EndTime = time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 0, 0, date.Location())
		}
		// fmt.Printf("Inserted 'study' activity from %s to %s\n", entry.StartTime.Format("15:04:05"), entry.EndTime.Format("15:04:05"))

		// remove the seconds part for start and end time
		entry.StartTime = entry.StartTime.Truncate(time.Minute)
		entry.EndTime = entry.EndTime.Truncate(time.Minute)

		dailyTracker := domain.DailyTracker{
			Activity:  "study", // Setting activity name to "study"
			StartTime: entry.StartTime,
			EndTime:   &entry.EndTime,
		}
		db.CreateDailyTracker(dailyTracker)
		// fmt.Printf("Inserted 'study' activity from %s to %s\n", entry.StartTime.Format("15:04:05"), entry.EndTime.Format("15:04:05"))
	}

	// if two entry gap is less than 10 minutes, and the activity all is "study", create a new entry activity name "study_break" with the gap time

	// Assuming 'entries' contains the sorted study activities for the date

	err = insertStudyBreak(day)
	if err != nil {
		fmt.Printf("Error inserting study break: %v\n", err)
		return
	}

	GetActivities(offset)

	fillMissingActivities(day)

	GetActivities(offset)

}

func insertStudyBreak(date string) error {

	// Fetch activities from daily_trackers for the given date
	var activities, err = db.SelectDailyTracker(date)
	if err != nil {
		return fmt.Errorf("error selecting daily tracker from day: %v", err)
	}

	// if two activity gap is
	//less than 10 minutes, and the two activity all is "study",
	// and  two activity is durartion all greater than 25 minutes
	//create a new entry activity name "study_break" with the gap time

	var lastEndTime time.Time
	var lastActivity string
	var lastDuration time.Duration
	for _, activity := range activities {
		if activity.Activity == defaultActivity {
			if activity.StartTime.Sub(lastEndTime) <= 10*time.Minute && lastActivity == defaultActivity && lastDuration >= pomodoroDuration {
				// create a new entry activity name "study_break" with the gap time
				dailyTracker := domain.DailyTracker{
					Activity:  "study_break", // Setting activity name to "study"
					StartTime: lastEndTime,
					EndTime:   &activity.StartTime,
				}
				db.CreateDailyTracker(dailyTracker)
			} else if activity.StartTime.Sub(lastEndTime) < 30*time.Minute && lastActivity == defaultActivity {
				// create a new entry activity name "study_break" with the gap time
				dailyTracker := domain.DailyTracker{
					Activity:  "study_distraction", // Setting activity name to "study"
					StartTime: lastEndTime,
					EndTime:   &activity.StartTime,
				}
				db.CreateDailyTracker(dailyTracker)
			}

		}
		lastEndTime = *activity.EndTime
		lastActivity = activity.Activity
		lastDuration = activity.EndTime.Sub(activity.StartTime)
	}

	return nil

}

func fillMissingActivities(date string) {
	fmt.Printf("\n%s  ", date)

fill:

	startTime, _ := time.ParseInLocation("2006-01-02", date, time.Local)
	// trunacate to minute
	startTime = startTime.Truncate(time.Minute)

	lastEndTime := startTime
	activities, _ := db.SelectDailyTracker(date)
	for _, activity := range activities {
		if activity.StartTime.After(lastEndTime) && activity.StartTime.Sub(lastEndTime) > time.Minute {
			// fmt.Printf("lastEndTime222: %s\n", lastEndTime.Format("15:04:05"))
			change := promptForActivity(lastEndTime, activity.StartTime)
			if change {
				goto fill
			}
		}
		lastEndTime = *activity.EndTime
	}
	fmt.Printf("lastEndTime: %s\n", lastEndTime.Format("15:04:05"))

	var endTime time.Time
	if date == time.Now().Format("2006-01-02") {
		endTime = time.Now()
	} else {
		endTime = startTime.Add(24*time.Hour - time.Minute)
	}
	endTime = endTime.Truncate(time.Minute)
	if endTime.After(lastEndTime) {
		change := promptForActivity(lastEndTime, endTime)
		if change {
			goto fill
		}

	}

}

func promptForActivity(startTime, endTime time.Time) bool {

	fmt.Printf("[%s - %s], total %v \n", startTime.Format("15:04:05"), endTime.Format("15:04:05"), endTime.Sub(startTime))

	var activity domain.DailyTracker
	reader := bufio.NewReader(os.Stdin)
	activity.Activity = getActivities(reader)
	var changeEndTime bool
	var start, end string

	fmt.Print("enter [hhmm-hhmm]  or use leave blank to use default: ")
	startEnd, _ := reader.ReadString('\n')
	startEnd = strings.TrimSpace(startEnd)
	startEnds := strings.Split(startEnd, "-")

	if len(startEnds) == 2 {
		start = startEnds[0]
		end = startEnds[1]
	} else {
		start = startEnds[0]
	}

	if start == "" {
		activity.StartTime = startTime
	} else {

		hourInt, err := strconv.Atoi(start[:2])
		if err != nil {
			log.Fatalf("error converting to int: %v", err)
		}
		minuteInt, err := strconv.Atoi(start[2:])
		if err != nil {
			log.Fatalf("error converting to int: %v", err)
		}

		activity.StartTime = time.Date(startTime.Year(), startTime.Month(), startTime.Day(), hourInt, minuteInt, 0, 0, time.Local)
		changeEndTime = true
	}

	if end == "" {
		activity.EndTime = &endTime
	} else {
		hourInt, err := strconv.Atoi(end[:2])
		if err != nil {
			log.Fatalf("error converting to int: %v", err)
		}
		minuteInt, err := strconv.Atoi(end[2:])
		if err != nil {
			log.Fatalf("error converting to int: %v", err)
		}
		// activity.EndTime = &(time.Date(endTime.Year(), endTime.Month(), endTime.Day(), hourInt, minuteInt, 0, 0, time.Local))

		endTime := time.Date(startTime.Year(), startTime.Month(), startTime.Day(), hourInt, minuteInt, 0, 0, time.Local)
		activity.EndTime = &endTime

		changeEndTime = true
	}

	if activity.EndTime.Before(activity.StartTime) {
		log.Fatalf("end time is before start time")
	}

	err := db.CreateDailyTracker(activity)
	if err != nil {
		log.Fatalf("error creating daily tracker: %v", err)
	}

	return changeEndTime
}

func getActivities(reader *bufio.Reader) string {

	choices, err := db.GetAllActivityName()
	if err != nil {
		log.Fatalf("error fetching distinct values %v", err)
	}
	sort.Strings(choices)

	for index, v := range choices {
		fmt.Printf("%d:%s, ", index+1, v)
	}

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return "play"
	}

	if num, err := strconv.Atoi(input); err == nil && num > 0 && num <= len(choices) {
		return choices[num-1]
	}

	for _, v := range choices {
		if strings.HasPrefix(v, input) {
			return v
		}
	}
	return input
}

func CrateActiveStart() {

	// crate a new activity which start time is current time, and end time is null, activity name is user input

	var activity domain.DailyTracker
	activity.StartTime = time.Now()

	// get activity name
	reader := bufio.NewReader(os.Stdin)
	activity.Activity = getActivities(reader)

	err := db.CreateDailyTracker(activity)
	if err != nil {

		log.Fatalf("error creating daily tracker: %v", err)
	}

}
