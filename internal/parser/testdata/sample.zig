const std = @import("std");
const mem = @import("std").mem;

/// A 2D point structure.
pub const Point = struct {
    x: f64,
    y: f64,

    /// Returns the distance from origin.
    pub fn distanceFromOrigin(self: Point) f64 {
        return @sqrt(self.x * self.x + self.y * self.y);
    }
};

/// Color options for rendering.
pub const Color = enum {
    red,
    green,
    blue,
};

/// Represents a network error.
const NetworkError = error{
    Timeout,
    ConnectionRefused,
    Unknown,
};

/// Adds two numbers together.
pub fn add(a: i32, b: i32) i32 {
    return a + b;
}

fn helper() void {
    // private function
}

/// Tagged union for results.
const Result = union(enum) {
    ok: i32,
    err: []const u8,
};

const MAX_SIZE: usize = 1024;

test "addition works" {
    try std.testing.expectEqual(@as(i32, 3), add(1, 2));
}

test "point distance" {
    const p = Point{ .x = 3.0, .y = 4.0 };
    try std.testing.expectApproxEqAbs(@as(f64, 5.0), p.distanceFromOrigin(), 0.001);
}
