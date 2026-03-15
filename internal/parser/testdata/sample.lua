local json = require("json")
local utils = require("utils")

--- A sample user class.
--- Handles user operations.
local User = {}
User.__index = User

--- Creates a new user instance.
function User:new(name, age)
  local self = setmetatable({}, User)
  self.name = name
  self.age = age
  return self
end

--- Returns a greeting string.
function User:greet(greeting)
  return greeting .. ", " .. self.name .. "!"
end

function User.find(id)
  return nil
end

local function helper()
  return 42
end

function globalFunc(x)
  return x * 2
end

function test_user_creation()
  -- test
end

function test_greeting()
  -- test
end
