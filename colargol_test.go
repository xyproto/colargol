package colargol

import (
	"testing"
	"time"
)

func TestLoadCalendarConfig(t *testing.T) {
	tests := []struct {
		locale string
	}{
		{"conf/nb_NO.tengo"},
		{"conf/en_US.tengo"},
		{"conf/tr_TR.tengo"},
		{"conf/sv_SE.tengo"},
		{"conf/de_DE.tengo"},
		{"conf/fr_FR.tengo"},
		{"conf/da_DK.tengo"},
	}

	for _, tt := range tests {
		t.Run(tt.locale, func(t *testing.T) {
			cal, err := NewCalendar(tt.locale)
			if err != nil {
				t.Fatalf("Failed to load calendar config for %s: %v", tt.locale, err)
			}
			if cal == nil {
				t.Fatalf("Calendar is nil for %s", tt.locale)
			}

			// Test some basic functions
			if got := cal.DayName(time.Monday); got == "" {
				t.Errorf("DayName(time.Monday) returned an empty string for %s", tt.locale)
			}
			if got := cal.MonthName(time.January); got == "" {
				t.Errorf("MonthName(time.January) returned an empty string for %s", tt.locale)
			}

			// Test red days
			red, desc, flag := cal.RedDay(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
			if !red {
				t.Errorf("RedDay returned false for a known red day for %s", tt.locale)
			}
			if desc == "" {
				t.Errorf("RedDay returned an empty description for %s", tt.locale)
			}
			if !flag {
				t.Errorf("RedDay returned false for flag day for %s", tt.locale)
			}

			// Test special days
			special, desc, flag := cal.SpecialDay(time.Date(2024, 2, 6, 0, 0, 0, 0, time.UTC))
			if !special {
				t.Errorf("SpecialDay returned false for a known special day for %s", tt.locale)
			}
			if desc == "" {
				t.Errorf("SpecialDay returned an empty description for %s", tt.locale)
			}
			if !flag {
				t.Errorf("SpecialDay returned false for flag day for %s", tt.locale)
			}

			// Test special periods (e.g., Easter)
			specialPeriod, name := cal.SpecialPeriod(time.Date(2024, 3, 31, 0, 0, 0, 0, time.UTC))
			if !specialPeriod {
				t.Errorf("SpecialPeriod returned false for a known special period for %s", tt.locale)
			}
			if name == "" {
				t.Errorf("SpecialPeriod returned an empty name for %s", tt.locale)
			}
		})
	}
}

func TestFormatFunctions(t *testing.T) {
	cal, err := NewCalendar("conf/nb_NO.tengo")
	if err != nil {
		t.Fatalf("Failed to load calendar config: %v", err)
	}
	if cal == nil {
		t.Fatalf("Calendar is nil")
	}

	// Test FormatDate
	expectedFormatDate := "17. okt"
	formattedDate := cal.FormatDate(17, 10, "okt")
	if formattedDate != expectedFormatDate {
		t.Errorf("FormatDate returned %q, want %q", formattedDate, expectedFormatDate)
	}

	// Test WeekString
	expectedWeekString := "Uke 7"
	weekString := cal.WeekString(7)
	if weekString != expectedWeekString {
		t.Errorf("WeekString returned %q, want %q", weekString, expectedWeekString)
	}

	// Test DayAndDate
	expectedDayAndDate := "mandag 24."
	dayAndDate := cal.DayAndDate("mandag", 24)
	if dayAndDate != expectedDayAndDate {
		t.Errorf("DayAndDate returned %q, want %q", dayAndDate, expectedDayAndDate)
	}
}
