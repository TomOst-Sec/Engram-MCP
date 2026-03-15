import java.util.List;
import java.util.Map;
import java.io.*;

/**
 * Represents a user in the system.
 * Handles user data and operations.
 */
public class User extends BaseEntity implements Serializable, Comparable<User> {

    private String name;
    private String email;
    protected int age;

    /**
     * Creates a new User with the given name and email.
     * @param name the user's name
     * @param email the user's email address
     */
    public User(String name, String email) {
        this.name = name;
        this.email = email;
    }

    /**
     * Returns the user's display name.
     * @return the formatted display name
     */
    public String getDisplayName() {
        return name + " <" + email + ">";
    }

    private void validateEmail(String email) {
        if (email == null || !email.contains("@")) {
            throw new IllegalArgumentException("Invalid email");
        }
    }

    protected List<String> getRoles() {
        return List.of("user");
    }

    @Override
    public int compareTo(User other) {
        return this.name.compareTo(other.name);
    }

    /**
     * Inner class for user preferences.
     */
    public static class Preferences {
        private Map<String, String> settings;

        public String getSetting(String key) {
            return settings.get(key);
        }
    }
}

/**
 * Defines the repository interface.
 */
interface UserRepository {
    User findById(long id);
    List<User> findAll();
    void save(User user);
}

enum UserRole {
    ADMIN,
    EDITOR,
    VIEWER
}

class UserServiceTest {
    @Test
    public void testCreateUser() {
        User user = new User("Alice", "alice@test.com");
        assert user.getDisplayName().contains("Alice");
    }

    @Test
    void testValidation() {
        // test validation logic
    }
}
