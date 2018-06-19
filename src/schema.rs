table! {
    devices (id) {
        id -> Int4,
        email -> Varchar,
        device_id -> Varchar,
        is_active -> Bool,
    }
}

table! {
    searches (id) {
        id -> Int4,
        email -> Varchar,
        sub -> Varchar,
        search -> Varchar,
    }
}

table! {
    user_info (email) {
        email -> Varchar,
        interval -> Float8,
    }
}

allow_tables_to_appear_in_same_query!(
    devices,
    searches,
    user_info,
);
