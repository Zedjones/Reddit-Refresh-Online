use super::schema::{searches, user_info, devices};

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

#[derive(Queryable)]
pub struct Device {
    pub id: i32, 
    pub email: String, 
    pub device_id: String,
    pub is_active: bool
}

#[derive(Insertable)]
#[table_name="devices"]
pub struct NewDevice<'a> {
    pub email: &'a str,
    pub device_id: &'a str, 
    pub is_active: bool
}