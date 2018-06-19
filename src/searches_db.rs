#[allow(dead_code)]
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

    pub fn get_searches(email: String) -> Vec<Search> {
        use schema::searches;
        let error = format!("Error loading searches for {}", email);
        let connection = connect();

        searches::table
            .filter(searches::email.eq(email))
            .load::<Search>(&connection)
            .expect(&error)

    }

    pub fn add_search(email: String, sub: String, search: String) -> Search {
        use schema::searches;
        use diesel::insert_into;
        let new_search = NewSearch { email:&email, sub:&sub, search:&search };
        let connection = connect();

        insert_into(searches::table)
            .values(&new_search)
            .get_result(&connection)
            .expect("Error saving new search")

    }

    #[allow(unused_imports)]
    pub fn delete_search(email_f: String, sub_f: String, search_f: String) 
    -> Result<(), String> {
        use schema::searches::dsl::*;
        use diesel::delete;

        let connection = connect();

        let num_deleted = delete(searches.filter(email.eq(email_f)
            .and(sub.eq(sub_f)).and(search.eq(search_f))))
            .execute(&connection)
            .expect("Error deleting post");

        match num_deleted {
            1 => Ok(()), 
            _ => Err("Invalid email, sub, or search to delete".to_string())
        }

    }

}