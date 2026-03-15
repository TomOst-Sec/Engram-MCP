use std::collections::HashMap;
use std::io::{self, Read, Write};

/// A user in the system.
/// Contains name and email fields.
#[derive(Debug, Clone)]
pub struct User {
    pub name: String,
    email: String,
    age: u32,
}

/// Represents possible errors.
pub enum AppError {
    NotFound(String),
    InvalidInput { field: String, reason: String },
    Internal,
}

/// Defines printable behavior.
pub trait Printable {
    /// Format as string for display.
    fn format(&self) -> String;
    fn print(&self);
}

impl Printable for User {
    fn format(&self) -> String {
        format!("{} <{}>", self.name, self.email)
    }

    fn print(&self) {
        println!("{}", self.format());
    }
}

impl User {
    /// Creates a new User.
    pub fn new(name: String, email: String, age: u32) -> Self {
        User { name, email, age }
    }

    pub fn greet(&self) -> String {
        format!("Hello, {}!", self.name)
    }
}

/// Processes input data and returns a result.
pub fn process_data<'a>(input: &'a str, count: usize) -> Result<Vec<&'a str>, AppError> {
    if input.is_empty() {
        return Err(AppError::InvalidInput {
            field: "input".to_string(),
            reason: "empty".to_string(),
        });
    }
    Ok(vec![input; count])
}

fn helper_function(x: i32, y: i32) -> i32 {
    x + y
}

macro_rules! create_map {
    ($($key:expr => $value:expr),*) => {
        {
            let mut map = HashMap::new();
            $(map.insert($key, $value);)*
            map
        }
    };
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_user_creation() {
        let user = User::new("Alice".to_string(), "alice@test.com".to_string(), 30);
        assert_eq!(user.name, "Alice");
    }

    #[test]
    fn test_process_data() {
        let result = process_data("hello", 3);
        assert!(result.is_ok());
    }
}
