import unittest
from calculator import add, subtract, multiply, divide


class TestCalculator(unittest.TestCase):

    def test_add(self):
        self.assertEqual(add(2, 3), 5)
        self.assertEqual(add(-1, 1), 0)
        self.assertEqual(add(0, 0), 0)

    def test_add_floats_and_negatives(self):
        self.assertAlmostEqual(add(0.1, 0.2), 0.3, places=7)
        self.assertEqual(add(-5.5, -4.5), -10.0)

    def test_subtract(self):
        self.assertEqual(subtract(5, 2), 3)
        self.assertEqual(subtract(2, 5), -3)
        self.assertEqual(subtract(0, 0), 0)

    def test_subtract_decimals(self):
        self.assertAlmostEqual(subtract(10.5, 3.2), 7.3, places=7)
        self.assertEqual(subtract(-2, -5), 3)

    def test_multiply(self):
        self.assertEqual(multiply(3, 4), 12)
        self.assertEqual(multiply(-2, 3), -6)
        self.assertEqual(multiply(0, 5), 0)

    def test_multiply_negative_numbers(self):
        self.assertEqual(multiply(-3, -3), 9)
        self.assertAlmostEqual(multiply(2.5, 4.0), 10.0)

    def test_divide(self):
        self.assertEqual(divide(10, 2), 5)
        self.assertAlmostEqual(divide(1, 3), 0.3333333, places=5)
        with self.assertRaises(ZeroDivisionError):
            divide(5, 0)

    def test_divide_negative_and_zero_numerator(self):
        self.assertEqual(divide(0, 5), 0)
        self.assertEqual(divide(-10, 2), -5)
        self.assertEqual(divide(10, -2), -5)


if __name__ == "__main__":
    unittest.main()
