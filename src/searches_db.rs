#[allow(dead_code)]
pub mod searches_db {

    use diesel::prelude::*;
    use diesel::pg::PgConnection;
    use dotenv::dotenv;
    use std::env;
    use models::{Search, NewSearch, UserInfo};

    /**
     * Function to connect to the database specified in .env
     * @return - a postgres database connection 
     */
    pub fn connect() -> PgConnection {
        
        dotenv().ok();

        let database_url = env::var("DATABASE_URL")
            .expect("DATABASE_URL must be set");
        PgConnection::establish(&database_url)
            .expect(&format!("Error connecting to {}", database_url))
    }

    /**
     * Get all the items in the searches table of the database which 
     * contain the requested email 
     * @param email - the email to filter searches by 
     * @return - a vector list containing search structs (id, email, sub, search)
     */
    pub fn get_searches(email: String) -> Vec<Search> {
        use schema::searches;
        let error = format!("Error loading searches for {}", email);
        let connection = connect();

        searches::table
            .filter(searches::email.eq(email))
            .load::<Search>(&connection)
            .expect(&error)
    }

    /**
     * Delete all queries in the searches table which contain the 
     * provided email and subreddit 
     * @param email - the email to delete queries for 
     * @param sub - the subreddit to delete queries for
     */
    pub fn delete_sub_searches(email_d: &str, sub_d: &str)
    -> Result<(), String> {
        use schema::searches::dsl::*;
        use diesel::delete;

        let connection = connect();

        let num_deleted = delete(searches.filter(email.eq(email_d)
            .and(sub.eq(sub_d))))
            .execute(&connection)
            .expect("Error deleting post");

        match num_deleted {
            0 => Err("Invalid email or sub to delete".to_string()),
            _ => Ok(())
        }
    }

    pub fn replace_sub_searches(email_d: &str, sub_d: &str, mut sub_searches: Vec<String>)
    -> Result<(), String> {
        use schema::searches::dsl::*;
        use diesel::{delete, insert_into};
        use diesel::dsl::{exists, select};

        let connection = connect();

        let _num_deleted = delete(searches.filter(email.eq(email_d)
            .and(sub.eq(sub_d))
            .and(search.eq_any(&sub_searches))))
            .execute(&connection)
            .expect("Error deleting searches");

        //FIXME: go back and make the Err statement do something meaningful
        sub_searches.retain(|ref sub_search| 
            match select(exists(searches.filter(
            search.eq(sub_search)))).get_result(&connection)
            {
                Ok(false) => true, 
                Ok(_) => false, 
                Err(_) => false
            });

        for search_d in &sub_searches {
			add_search(&email_d, &sub_d, &search_d);
        }

        Ok(())
    }

    /**
     * Add a search with the provided email, subreddit, and search term 
     * @param email - the email to use for this item 
     * @param sub - the subreddit to use for this item
     * @param search - the search term to use for this item 
     * @return - the search object that was added
     */
    pub fn add_search(email: &str, sub: &str, search: &str) -> Search {
        use schema::searches;
        use diesel::insert_into;
        let new_search = NewSearch { email:&email, sub:&sub, search:&search, last_res_url:&"" };
        let connection = connect();

        insert_into(searches::table)
            .values(&new_search)
            .get_result(&connection)
            .expect("Error saving new search")

    }

    /**
     * Deletes a query with the given email, subreddit, and search term 
     * @param email_f - the email of query to delete
     * @param sub_f - the subreddit of query to delete
     * @param search_f - the search of query to delete 
     * @return - Either ok() or an error about why the deletion failed 
     */
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

    pub fn get_interval(email: &str) -> f64 {
        use schema::user_info;
        let error = format!("Error loading interval for {}", email);
        let connection = connect();

        let weird_vec = user_info::table
            .filter(user_info::email.eq(email))
            .load::<UserInfo>(&connection)
            .expect(&error);

        weird_vec[0].interval
    }

    pub fn update_last_res(email_f: String, sub_f: String, search_f: String, new_res: String) {
        use schema::searches::dsl::*;
        use diesel::update;
        let connection = connect();

        update(searches.filter(email.eq(email_f)
            .and(sub.eq(sub_f)).and(search.eq(search_f))))
            .set(last_res_url.eq(new_res))
            .execute(&connection)
            .expect("Error updating url");

        ()
    }

}