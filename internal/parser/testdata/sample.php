<?php
namespace App\Models;

use App\Contracts\Displayable;
use App\Traits\HasTimestamps;

/**
 * Represents a user in the system.
 * Handles authentication and profile management.
 */
class User extends BaseModel implements Displayable {
    use HasTimestamps;

    public string $name;
    private int $age;

    /**
     * Creates a new User instance.
     */
    public function __construct(string $name, int $age) {
        $this->name = $name;
        $this->age = $age;
    }

    /**
     * Returns a greeting message for the user.
     */
    public function greet(string $greeting): string {
        return "{$greeting}, {$this->name}!";
    }

    private function validateEmail(): void {
        // validation logic
    }

    protected function getDisplayName(): string {
        return $this->name;
    }

    public function testSomething(): void {
        // test method
    }
}

/**
 * Defines the display contract.
 */
interface Displayable {
    public function render(bool $verbose): string;
    public function getFormat(): string;
}

trait HasTimestamps {
    public function createdAt(): \DateTime {
        return new \DateTime();
    }
}

/**
 * A standalone helper function.
 */
function helperFunc(int $x): int {
    return $x * 2;
}

enum UserRole: string {
    case Admin = 'admin';
    case Editor = 'editor';
    case Viewer = 'viewer';
}

class UserTest extends TestCase {
    public function testUserCreation(): void {
        // test
    }

    public function testGreeting(): void {
        // test
    }

    public function helperMethod(): void {
        // not a test
    }
}
