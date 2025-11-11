package main

// versionMap translates the raw major version integer
// into the common Adobe product name.
var versionMap = map[uint32]string{
	3:  "CS",
	4:  "CS2",
	5:  "CS3",
	6:  "CS4",
	7:  "CS5",
	8:  "CS6",
	9:  "CC",
	10: "CC 2014",
	11: "CC 2015",
	12: "CC 2017",
	13: "CC 2018",
	14: "CC 2019",
	15: "2020",
	16: "2021",
	17: "2022",
	18: "2023",
	19: "2024",
	20: "2025",
	21: "2026",
	//being optimistic here...
	22: "2027",
	23: "2028",
}

// List of keywords to ignore in the application path
var ignoreKeywords = []string{"server", "debug", "prerelease", "beta"}

// reverseVersionMap translates the product name (e.g., "2024")
// back to the major version number (e.g., 19).
// We'll build this automatically from versionMap.
var reverseVersionMap = make(map[string]uint32)

// init() runs once, automatically, when the program starts.
func init() {
	for v, name := range versionMap {
		reverseVersionMap[name] = v
	}
}