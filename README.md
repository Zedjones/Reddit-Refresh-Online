Reddit-Refresh-Online
=======
A web application that allows users to receive Pushbullet notifications whenever a new result shows up for a search query on a given subreddit. A user can monitor as many queries on as many subreddits as they wish, and each one is monitored in a separate coroutine using Go's goroutines. 
The project uses:
 - The Pushbullet OAuth for login, with the token being stored in the database
 - The Pushbullet API to send notifications to users, on the devices they specify
 - PostgreSQL for the database
 - The Reddit API to get new search results
 - Materialize CSS and Vanilla JS for the frontend
 - Golang with the Echo web framework for the backend
