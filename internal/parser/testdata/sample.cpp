#include <iostream>
#include <vector>

using namespace std;

namespace math {

/// A 2D vector class for math operations.
class Vector2D : public Printable {
public:
    /// Constructs a new 2D vector.
    Vector2D(double x, double y) : x_(x), y_(y) {}
    ~Vector2D() = default;

    /// Returns the magnitude of the vector.
    double magnitude() const {
        return sqrt(x_ * x_ + y_ * y_);
    }

    virtual void print() const override {
        cout << x_ << ", " << y_ << endl;
    }

    Vector2D operator+(const Vector2D& other) const {
        return Vector2D(x_ + other.x_, y_ + other.y_);
    }

private:
    double x_;
    double y_;
};

/// Returns the maximum of two values.
template<typename T>
T max_val(T a, T b) {
    return (a > b) ? a : b;
}

enum class Direction { North, South, East, West };

/// A helper function outside the class.
int helperFunc(int x) {
    return x * 2;
}

} // namespace math
