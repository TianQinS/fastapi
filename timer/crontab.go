// Crontab like the executable crontab in unix.
// It's configuration cmd is as follows too, and it's work mechanism is check all crontab every minute.
// *      0     1    2    3    4
// *      *     *    *    *    *
// *      -     -    -    -    -
// *      |     |    |    |    |
// *      |     |    |    |    +----- day of week (0 - 6) (Sunday=0)
// *      |     |    |    +------- month (1 - 12)
// *      |     |    +--------- day of month (1 - 31)
// *      |     +----------- hour (0 - 23)
// *      +------------- minute (0 - 59)
package timer

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	CRONTAB_ATOMS_LEN = 5
)

var (
	crontabFormat = [][]int{
		[]int{0, 59},
		[]int{0, 23},
		[]int{1, 31},
		[]int{1, 12},
		[]int{1, 7},
	}
	lock             = new(sync.Mutex)
	cancelledHandles = []Handle{}
	entries          = map[Handle]*entry{}
	// for quick match like timewheel.
	entryWheelMap = map[int]map[Handle]*entry{}
	nextHandle    = Handle(1)
	// for qc test.
	qcTimer     *Timer
	qcDeltaTime time.Duration
)

// Handle is the type of return value of Register, can be used to cancel the register
type Handle int

type entry struct {
	// the executable crontab command
	crontab string
	// the memo field
	info                                string
	minute, hour, day, month, dayofweek int64
	cb                                  interface{}
	params                              []interface{}
}

func (this *entry) getInfo() string {
	return fmt.Sprintf("%s | %s", this.crontab, this.info)
}

// Parse one atom in Crontab, return the valid values map.
func (this *entry) parse(atom string, seq int) int64 {
	divisor, result, start, end := 1, int64(0), crontabFormat[seq][0], crontabFormat[seq][1]
	values := make([]int, 0)

	// The divisor of a time str.
	if strings.Contains(atom, "/") {
		tmp := strings.Split(atom, "/")
		atom = tmp[0]
		if _divisor, err := strconv.Atoi(tmp[1]); err == nil {
			divisor = _divisor
		}
	}
	if strings.Contains(atom, "-") {
		//  A continuum like 2-7
		tmp := strings.Split(atom, "-")
		_start, _ := strconv.Atoi(tmp[0])
		_end, _ := strconv.Atoi(tmp[1])
		if _start >= start && _end <= end {
			start = _start
			end = _end
			atom = "*"
		}
	} else if strings.Contains(atom, ",") {
		// Discrete numbers like 4,6
		tmp := strings.Split(atom, ",")
		for _, t := range tmp {
			if val, err := strconv.Atoi(t); err == nil {
				values = append(values, val)
			}
		}
	} else {
		// Single number like 5
		if val, err := strconv.Atoi(atom); err == nil {
			values = append(values, val)
		}
	}

	if atom == "*" {
		// All possible values.
		for i := start; i <= end; i++ {
			values = append(values, i)
		}
	}
	for _, val := range values {
		if val%divisor != 0 {
			continue
		}
		// fmt.Printf("%d ", val)
		result |= (1 << uint(val))
	}
	// fmt.Println("")

	return result
}

//  Check all crontab's map data every minute.
func (this *entry) match(minute int, hour int, day int, month time.Month, weekday time.Weekday) bool {
	if this.minute&(1<<uint(minute)) == 0 {
		return false
	}
	if this.hour&(1<<uint(hour)) == 0 {
		return false
	}
	if this.day&(1<<uint(day)) == 0 {
		return false
	}
	if this.month&(1<<uint(month)) == 0 {
		return false
	}

	// Special treatment for week
	week := uint(weekday)
	if week == 0 {
		week = 7
	}
	if this.dayofweek&(1<<week) == 0 {
		return false
	}
	return true
}

// Parse all five atoms in Crontab, update the valid values map.
func (this *entry) parseValidAtoms() {
	cmds := strings.Split(this.crontab, " ")
	if len(cmds) != CRONTAB_ATOMS_LEN {
		fmt.Fprintf(os.Stderr, "Crontab %s error\n", cmds)
	}
	this.minute = this.parse(cmds[0], 0)
	this.hour = this.parse(cmds[1], 1)
	this.day = this.parse(cmds[2], 2)
	this.month = this.parse(cmds[3], 3)
	this.dayofweek = this.parse(cmds[4], 4)
}

func (this *entry) genWheelKey() []int {
	keys := make([]int, 0, 1)
	for i := 0; i < 60; i++ {
		if this.minute&(1<<uint(i)) != 0 {
			for j := 0; j < 24; j++ {
				if this.hour&(1<<uint(j)) != 0 {
					keys = append(keys, j*60+i)
				}
			}
		}
	}
	return keys
}

func addWheelMap(h Handle, slot *entry) {
	keys := slot.genWheelKey()
	for _, key := range keys {
		entryWheelMap[key][h] = slot
	}
}

func clearWheelMap(h Handle, slot *entry) {
	keys := slot.genWheelKey()
	for _, key := range keys {
		delete(entryWheelMap[key], h)
	}
}

// Register a callack which will be executed when time condition is satisfied
func AddCrontab(crontab, info string, cb interface{}, params ...interface{}) Handle {
	h := genNextHandle()
	slot := &entry{
		crontab:   crontab,
		info:      info,
		minute:    0,
		hour:      0,
		day:       0,
		month:     0,
		dayofweek: 0,
		cb:        cb,
		params:    params,
	}
	slot.parseValidAtoms()
	addWheelMap(h, slot)
	defer lock.Unlock()
	lock.Lock()
	entries[h] = slot
	return h
}

// Unregister a registered crontab handle
func (h Handle) Cancel() {
	defer lock.Unlock()
	lock.Lock()
	cancelledHandles = append(cancelledHandles, h)
}

func unregisterCancelledHandles() {
	for _, h := range cancelledHandles {
		if slot, ok := entries[h]; ok {
			clearWheelMap(h, slot)
			delete(entries, h)
		}
	}
	cancelledHandles = nil
}

func check() {
	unregisterCancelledHandles()
	now := GetQcTime()
	dayofweek, month, day, hour, minute := now.Weekday(), now.Month(), now.Day(), now.Hour(), now.Minute()
	key := hour*60 + minute
	for _, slot := range entryWheelMap[key] {
		if slot.match(minute, hour, day, month, dayofweek) {
			GPost.PutJob(_TIMER_JOB_GROUP, slot.cb, slot.params...)
		}
	}
}

func genNextHandle() (h Handle) {
	defer lock.Unlock()
	lock.Lock()
	h, nextHandle = nextHandle, nextHandle+1
	return
}

// Initialize crontab module, called by mtimer.
func Initialize(first bool) {
	now := GetQcTime()
	sec := now.Second()
	d := time.Second*time.Duration(60+1-sec) - time.Nanosecond*time.Duration(now.Nanosecond())
	if !first {
		d += time.Minute
	}
	AddCallback(d, func() {
		if qcTimer != nil {
			qcTimer.Cancel()
		}
		qcTimer = AddTimer(time.Minute, check)
		check()
	})
}

// Set time for for purposes of testing with lazy executing.
// The check time point in this mode is not accurate.
func SetQcTime(stime string) error {
	target, err := time.ParseInLocation(TIME_FORMAT, stime, time.Local)
	if err == nil {
		qcDeltaTime = target.Sub(time.Now())
		Initialize(false)
	}
	return err
}

func GetQcTime() time.Time {
	return time.Now().Add(qcDeltaTime)
}

func ClearQcTime() {
	qcDeltaTime = time.Duration(0) // int64
	Initialize(false)
}

// For self-test.
func TestCrontab(stime string) (error, map[Handle]string) {
	now, err := time.ParseInLocation(TIME_FORMAT, stime, time.Local)
	info := make(map[Handle]string, len(entries))
	if err == nil {
		unregisterCancelledHandles()
		dayofweek, month, day, hour, minute := now.Weekday(), now.Month(), now.Day(), now.Hour(), now.Minute()
		key := hour*60 + minute
		for _, slot := range entryWheelMap[key] {
			if slot.match(minute, hour, day, month, dayofweek) {
				GPost.PutJob(_TIMER_JOB_GROUP, slot.cb, slot.params...)
			}
		}
	}
	for handle, entry := range entries {
		info[handle] = entry.getInfo()
	}
	return err, info
}

// Initialize crontab wheel map.
func startCrontab() {
	qcTimer = nil
	qcDeltaTime = time.Duration(0)
	entryWheelMap = make(map[int]map[Handle]*entry, 24*60)
	for i := 0; i < 60; i++ {
		for j := 0; j < 24; j++ {
			entryWheelMap[60*j+i] = make(map[Handle]*entry)
		}
	}
	Initialize(true)
}
