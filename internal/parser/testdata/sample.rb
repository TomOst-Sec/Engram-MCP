# A sample Ruby file for testing the parser

require 'json'
require_relative 'helper'

# The User class handles user operations.
# It extends BaseModel for persistence.
class User < BaseModel
  include Printable
  extend ClassMethods

  attr_reader :name, :email
  attr_accessor :age

  ROLE_ADMIN = 'admin'
  MAX_AGE = 150

  # Creates a new user instance.
  def initialize(name, email, age)
    @name = name
    @email = email
    @age = age
  end

  # Returns a greeting string for the user.
  def greet(greeting)
    "#{greeting}, #{@name}!"
  end

  # Finds a user by ID.
  def self.find(id)
    # class method lookup
  end

  def self.create(attrs)
    # another class method
  end

  private

  def validate_email
    raise "Invalid" if @email.nil?
  end

  def test_greeting
    # test method
  end

  def test_validation
    # another test
  end
end

# Serializable module for JSON conversion.
module Serializable
  def to_json
    # serialize to JSON
  end

  def from_json(data)
    # deserialize from JSON
  end
end
