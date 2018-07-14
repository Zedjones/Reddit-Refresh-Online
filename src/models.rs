use super::schema::searches;
use super::schema::user_info;

#[derive(Queryable)]
pub struct Search {
    pub id: i32,
    pub email: String,
    pub sub: String,
    pub search: String,
    pub last_res_url: String
}

#[derive(Insertable)]
#[table_name="searches"]
pub struct NewSearch<'a> {
    pub email: &'a str,
    pub sub: &'a str,
    pub search: &'a str,
    pub last_res_url: &'a str
}

#[derive(Queryable)]
pub struct UserInfo {
    pub email: String,
    pub interval: f64
}

#[derive(Insertable)]
#[table_name="user_info"]
pub struct NewUserInfo<'a> {
    pub email: &'a str,
    pub interval: f64
}