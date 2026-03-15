import kotlin.collections.List
import kotlin.io.println

/**
 * Represents a user in the system.
 * Handles user data and operations.
 */
open class User(
    val name: String,
    private val email: String,
    protected var age: Int
) : BaseEntity(), Serializable {

    /**
     * Returns the formatted display name.
     */
    fun getDisplayName(): String {
        return "$name <$email>"
    }

    private fun validateEmail(): Boolean {
        return email.contains("@")
    }

    protected fun getRoles(): List<String> {
        return listOf("user")
    }

    companion object {
        fun create(name: String): User {
            return User(name, "", 0)
        }
    }
}

/**
 * A data class for user preferences.
 */
data class UserPreferences(
    val theme: String,
    val language: String,
    val notifications: Boolean
)

/**
 * Repository interface for user operations.
 */
interface UserRepository {
    fun findById(id: Long): User?
    fun findAll(): List<User>
    fun save(user: User)
}

sealed class Result<out T> {
    data class Success<T>(val data: T) : Result<T>()
    data class Error(val message: String) : Result<Nothing>()
    object Loading : Result<Nothing>()
}

object AppConfig {
    val version = "1.0.0"
    fun getEnvironment(): String = "production"
}

/**
 * Processes input data and returns results.
 */
fun processData(input: String, count: Int): List<String> {
    return List(count) { input }
}

fun String.toSlug(): String {
    return this.lowercase().replace(" ", "-")
}

internal fun helperFunction(x: Int, y: Int): Int {
    return x + y
}

class UserTests {
    @Test
    fun testCreateUser() {
        val user = User("Alice", "alice@test.com", 30)
        assert(user.name == "Alice")
    }

    @Test
    fun testDisplayName() {
        val user = User("Bob", "bob@test.com", 25)
        assert(user.getDisplayName().contains("Bob"))
    }
}
