import Foundation
import UIKit

/// A user in the system.
/// Manages user profile data.
class User: BaseEntity, Codable, Equatable {
    var name: String
    private var email: String
    let id: Int

    /// Creates a new User instance.
    init(name: String, email: String, id: Int) {
        self.name = name
        self.email = email
        self.id = id
    }

    /// Returns the display name.
    func getDisplayName() -> String {
        return "\(name) <\(email)>"
    }

    private func validateEmail() -> Bool {
        return email.contains("@")
    }

    static func create(name: String) -> User {
        return User(name: name, email: "", id: 0)
    }
}

/// Represents a point in 2D space.
struct Point: Hashable {
    var x: Double
    var y: Double

    func distance(to other: Point) -> Double {
        let dx = x - other.x
        let dy = y - other.y
        return (dx * dx + dy * dy).squareRoot()
    }
}

/// Defines printable behavior.
protocol Printable {
    func format() -> String
    func print()
}

/// Possible app states.
enum AppState {
    case loading
    case ready(data: String)
    case error(Error)
}

extension User: Printable {
    func format() -> String {
        return getDisplayName()
    }

    func print() {
        Swift.print(format())
    }
}

/// Processes input data and returns results.
func processData(input: String, count: Int) -> [String] {
    return Array(repeating: input, count: count)
}

func helperFunction(_ x: Int, _ y: Int) -> Int {
    return x + y
}

class UserTests: XCTestCase {
    func testCreateUser() {
        let user = User(name: "Alice", email: "alice@test.com", id: 1)
        XCTAssertEqual(user.name, "Alice")
    }

    func testDisplayName() {
        let user = User(name: "Bob", email: "bob@test.com", id: 2)
        XCTAssertTrue(user.getDisplayName().contains("Bob"))
    }
}
