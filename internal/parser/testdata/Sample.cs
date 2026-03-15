using System;
using System.Collections.Generic;
using static System.Console;

namespace Engram.Sample
{
    /// <summary>
    /// Represents a user in the system.
    /// Handles authentication and profile management.
    /// </summary>
    public class User : BaseEntity, IDisplayable
    {
        /// <summary>
        /// Gets or sets the user's name.
        /// </summary>
        public string Name { get; set; }

        public int Age { get; private set; }

        private readonly string _email;

        /// <summary>
        /// Creates a new User instance.
        /// </summary>
        public User(string name, string email, int age)
        {
            Name = name;
            _email = email;
            Age = age;
        }

        /// <summary>
        /// Returns a greeting message for the user.
        /// </summary>
        public string Greet(string greeting)
        {
            return $"{greeting}, {Name}!";
        }

        private void ValidateEmail()
        {
            if (string.IsNullOrEmpty(_email))
                throw new ArgumentException("Email required");
        }

        protected internal virtual string GetDisplayName()
        {
            return Name;
        }

        /// <summary>
        /// Retrieves items of the specified type.
        /// </summary>
        public List<T> GetItems<T>(Func<T, bool> predicate) where T : IComparable
        {
            return new List<T>();
        }

        public class InnerSettings
        {
            public string Theme { get; set; }
        }
    }

    /// <summary>
    /// Defines the display contract.
    /// </summary>
    public interface IDisplayable
    {
        string GetDisplayName();
        void Render(bool verbose);
    }

    public struct Point : IComparable<Point>
    {
        public double X { get; set; }
        public double Y { get; set; }

        public int CompareTo(Point other)
        {
            return X.CompareTo(other.X);
        }
    }

    public enum UserRole
    {
        Admin,
        Editor,
        Viewer,
        Guest
    }

    public record UserRecord(string Name, string Email, int Age);

    public abstract class BaseEntity
    {
        public Guid Id { get; set; }
        public DateTime CreatedAt { get; set; }

        public abstract string Describe();
    }

    public class UserTests
    {
        [Test]
        public void TestUserCreation()
        {
            var user = new User("Alice", "alice@example.com", 30);
            Assert.AreEqual("Alice", user.Name);
        }

        [Fact]
        public void UserGreetReturnsCorrectMessage()
        {
            var user = new User("Bob", "bob@example.com", 25);
            Assert.Equal("Hello, Bob!", user.Greet("Hello"));
        }

        [TestMethod]
        public void ValidateEmailThrowsOnEmpty()
        {
            // test body
        }
    }
}
