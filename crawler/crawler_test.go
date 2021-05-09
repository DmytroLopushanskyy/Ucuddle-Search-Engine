package main


// ---------- after adding new links in Task Manager
// LAST_LINK_ID_KEY_ELASTIC in INDEX_CONFIG_ELASTIC must increase on length of added links

// ---------- during work of crawler in the field "parsed" must appear new true values

// ---------- after end of work of Task Manager next values must change
// - in the index for links all values in the field "taken" must be true
// - field "parsed" in some places false, but in the majority must be true

// ---------- Test iteration over indexes_names in INDEXES_ELASTIC_LINKS.split()
// - must correct iterate over these indexes_names and stop when reached the end correctly
// - if there is no such index_name, so should skip it
// - test different numbers of links in indexes
// - test when stop crawler and rerun -- must start from the previous unparsed domains
// - test if really parse links which were taken: true, parsed: false
