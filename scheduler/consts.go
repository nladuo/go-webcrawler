package scheduler

const (
	//if the tasks or results length more than 250,
	//serialize the task and store it into sql database
	store_to_sql_count int = 250
	store_count        int = 100 //store 100 tasks or results to database

	//if the tasks or results length less than 50,
	//get 100 tasks or results from sql database
	extract_from_sql_count int = 50
	extract_count          int = 100 //extract 100 tasks or results to memeory
	// the channal's buffer size
	chan_buffer_size int = 300
)
