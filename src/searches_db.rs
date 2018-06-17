pub mod searches_db {

    use diesel::prelude::*;
    use diesel::pg::PgConnection;
    use dotenv::dotenv;
    use std::env;
    use models::{Search, NewSearch};

    pub fn connect() -> PgConnection {
        
        dotenv().ok();

        let database_url = env::var("DATABASE_URL")
            .expect("DATABASE_URL must be set");
        PgConnection::establish(&database_url)
            .expect(&format!("Error connecting to {}", database_url))
    }

    pub fn get_searches(username: String) -> Vec<Search> {
        use schema::searches;
        let error = format!("Error loading searches for {}", username);
        let connection = connect();

        searches::table
            .filter(searches::username.eq(username))
            .load::<Search>(&connection)
            .expect(&error)

    }

    pub fn add_search(username: String, sub: String, search: String) -> Search {
        use schema::searches;
        use diesel::insert_into;
        let new_search = NewSearch { username:&username, sub:&sub, search:&search };
        let connection = connect();

        insert_into(searches::table)
            .values(&new_search)
            .get_result(&connection)
            .expect("Error saving new search")

    }

    pub fn delete_search(username: String, sub: String, search: String) {
        use schema::searches;
        let connection = connect();

        //TODO add deletion mechanism for only an exact match
    }

}