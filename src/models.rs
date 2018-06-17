use super::schema::searches;

#[derive(Queryable)]
pub struct Search {
    pub id: i32,
    pub email: String,
    pub sub: String,
    pub search: String,
}

#[derive(Insertable)]
#[table_name="searches"]
pub struct NewSearch<'a> {
    pub email: &'a str,
    pub sub: &'a str,
    pub search: &'a str
}