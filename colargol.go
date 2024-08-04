package colargol

import (
	"os"
	"time"

	"github.com/d5/tengo/v2"
)

// Calendar provides a common interface for calendars of all languages and locales
type Calendar interface {
	DayName(time.Weekday) string
	MonthName(time.Month) string
	RedDay(time.Time) (bool, string, bool)
	SpecialDay(time.Time) (bool, string, bool)
	SpecialPeriod(time.Time) (bool, string)
	NormalDay() string
	MondayFirst() bool
	FormatDate(int, int, string) string
	WeekString(int) string
	DayAndDate(string, int) string
}

// CustomCalendar implements the Calendar interface based on Tengo configuration
type CustomCalendar struct {
	dayNames       []string
	monthNames     []string
	holidays       []Holiday
	specialDays    []SpecialDay
	specialPeriods []SpecialPeriod
	normalDayDesc  string
	mondayFirst    bool
	script         *tengo.Script
}

// Holiday represents a holiday entry in the config
type Holiday struct {
	Date        string
	Description string
	FlagDay     bool
}

// SpecialDay represents a special day entry in the config
type SpecialDay struct {
	Date        string
	Description string
	FlagDay     bool
}

// SpecialPeriod represents a special period entry in the config
type SpecialPeriod struct {
	Name        string
	TengoScript string
}

func (cc *CustomCalendar) DayName(day time.Weekday) string {
	if int(day) < len(cc.dayNames) {
		return cc.dayNames[day]
	}
	return day.String()
}

func (cc *CustomCalendar) MonthName(month time.Month) string {
	if int(month)-1 < len(cc.monthNames) {
		return cc.monthNames[month-1]
	}
	return month.String()
}

func (cc *CustomCalendar) RedDay(date time.Time) (bool, string, bool) {
	for _, holiday := range cc.holidays {
		holidayDate, _ := time.Parse("2006-01-02", holiday.Date)
		if date.Equal(holidayDate) {
			return true, holiday.Description, holiday.FlagDay
		}
	}
	return false, "", false
}

func (cc *CustomCalendar) SpecialDay(date time.Time) (bool, string, bool) {
	for _, special := range cc.specialDays {
		specialDate, _ := time.Parse("2006-01-02", special.Date)
		if date.Equal(specialDate) {
			return true, special.Description, special.FlagDay
		}
	}
	return false, "", false
}

func (cc *CustomCalendar) SpecialPeriod(date time.Time) (bool, string) {
	for _, period := range cc.specialPeriods {
		script := tengo.NewScript([]byte(period.TengoScript))
		script.Add("year", int64(date.Year()))
		script.Add("month", int64(date.Month()))
		script.Add("day", int64(date.Day()))
		registerFunctions(script)

		compiled, err := script.Run()
		if err != nil {
			return false, ""
		}

		result := compiled.Get("result")
		if result == nil || result.IsFalsy() {
			return false, ""
		}

		return true, period.Name
	}
	return false, ""
}

func (cc *CustomCalendar) NormalDay() string {
	return cc.normalDayDesc
}

func (cc *CustomCalendar) MondayFirst() bool {
	return cc.mondayFirst
}

func (cc *CustomCalendar) FormatDate(day, month int, monthAbbrev string) string {
	cc.script.Add("day", day)
	cc.script.Add("month", month)
	cc.script.Add("month_abbrev", monthAbbrev)
	result, err := cc.script.Run()
	if err != nil {
		return ""
	}
	return result.Get("result").String()
}

func (cc *CustomCalendar) WeekString(week int) string {
	cc.script.Add("week", week)
	result, err := cc.script.Run()
	if err != nil {
		return ""
	}
	return result.Get("result").String()
}

func (cc *CustomCalendar) DayAndDate(dayName string, day int) string {
	cc.script.Add("day_name", dayName)
	cc.script.Add("day", day)
	result, err := cc.script.Run()
	if err != nil {
		return ""
	}
	return result.Get("result").String()
}

// NewCalendar creates a new calendar based on the given Tengo configuration file.
func NewCalendar(tengoConfigPath string) (Calendar, error) {
	file, err := os.ReadFile(tengoConfigPath)
	if err != nil {
		return nil, err
	}

	script := tengo.NewScript(file)
	compiled, err := script.Run()
	if err != nil {
		return nil, err
	}

	cc := &CustomCalendar{
		dayNames:       getStringList(compiled.Get("day_names")),
		monthNames:     getStringList(compiled.Get("month_names")),
		holidays:       getHolidays(compiled.Get("holidays")),
		specialDays:    getSpecialDays(compiled.Get("special_days")),
		specialPeriods: getSpecialPeriods(compiled.Get("special_periods")),
		normalDayDesc:  compiled.Get("normal_day_desc").String(),
		mondayFirst:    compiled.Get("monday_first").Bool(),
		script:         script,
	}

	return cc, nil
}

func getStringList(array tengo.Object) []string {
	arr := array.(*tengo.Array)
	var list []string
	for _, item := range arr.Value {
		list = append(list, item.(*tengo.String).Value)
	}
	return list
}

func getHolidays(array tengo.Object) []Holiday {
	arr := array.(*tengo.Array)
	var holidays []Holiday
	for _, item := range arr.Value {
		mapItem := item.(*tengo.Map)
		holidays = append(holidays, Holiday{
			Date:        mapItem.Value["date"].(*tengo.String).Value,
			Description: mapItem.Value["description"].(*tengo.String).Value,
			FlagDay:     !mapItem.Value["flag_day"].IsFalsy(),
		})
	}
	return holidays
}

func getSpecialDays(array tengo.Object) []SpecialDay {
	arr := array.(*tengo.Array)
	var specialDays []SpecialDay
	for _, item := range arr.Value {
		mapItem := item.(*tengo.Map)
		specialDays = append(specialDays, SpecialDay{
			Date:        mapItem.Value["date"].(*tengo.String).Value,
			Description: mapItem.Value["description"].(*tengo.String).Value,
			FlagDay:     !mapItem.Value["flag_day"].IsFalsy(),
		})
	}
	return specialDays
}

func getSpecialPeriods(array tengo.Object) []SpecialPeriod {
	arr := array.(*tengo.Array)
	var specialPeriods []SpecialPeriod
	for _, item := range arr.Value {
		mapItem := item.(*tengo.Map)
		specialPeriods = append(specialPeriods, SpecialPeriod{
			Name:        mapItem.Value["name"].(*tengo.String).Value,
			TengoScript: mapItem.Value["tengo_script"].(*tengo.String).Value,
		})
	}
	return specialPeriods
}

// Register Go functions in Tengo
func registerFunctions(script *tengo.Script) {
	script.Add("is_easter", isEasterFunc)
	script.Add("add_days", addDaysFunc)
}

var isEasterFunc = &tengo.UserFunction{
	Name: "is_easter",
	Value: func(args ...tengo.Object) (tengo.Object, error) {
		if len(args) != 3 {
			return nil, tengo.ErrWrongNumArguments
		}
		year, _ := tengo.ToInt64(args[0])
		month, _ := tengo.ToInt64(args[1])
		day, _ := tengo.ToInt64(args[2])
		return tengo.TrueValue, isEaster(int(year), int(month), int(day))
	},
}

var addDaysFunc = &tengo.UserFunction{
	Name: "add_days",
	Value: func(args ...tengo.Object) (tengo.Object, error) {
		if len(args) != 4 {
			return nil, tengo.ErrWrongNumArguments
		}
		year, _ := tengo.ToInt64(args[0])
		month, _ := tengo.ToInt64(args[1])
		day, _ := tengo.ToInt64(args[2])
		days, _ := tengo.ToInt64(args[3])
		newDate := addDays(int(year), int(month), int(day), int(days))
		return &tengo.Array{
			Value: []tengo.Object{
				&tengo.Int{Value: int64(newDate.Year())},
				&tengo.Int{Value: int64(newDate.Month())},
				&tengo.Int{Value: int64(newDate.Day())},
			},
		}, nil
	},
}

// Go implementation of the Easter calculation
func isEaster(year, month, day int) bool {
	a := year % 19
	b := year / 100
	c := year % 100
	d := b / 4
	e := b % 4
	f := (b + 8) / 25
	g := (b - f + 1) / 3
	h := (19*a + b - d - g + 15) % 30
	i := c / 4
	k := c % 4
	l := (32 + 2*e + 2*i - h - k) % 7
	m := (a + 11*h + 22*l) / 451
	n := h + l - 7*m + 114
	easterMonth := n / 31
	easterDay := (n % 31) + 1
	return month == easterMonth && day == easterDay
}

// Go implementation to add days to a given date
func addDays(year, month, day, days int) time.Time {
	date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return date.AddDate(0, 0, days)
}
