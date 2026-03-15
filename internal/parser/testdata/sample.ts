import { Request, Response } from 'express';
import * as crypto from 'crypto';
import defaultExport from './utils';

/**
 * Represents a user in the system.
 */
interface UserInterface {
  name: string;
  email: string;
  greet(greeting: string): string;
}

/** Status type alias */
type Status = 'active' | 'inactive' | 'banned';

/** Available roles */
enum Role {
  Admin = 'admin',
  User = 'user',
  Guest = 'guest',
}

/**
 * User class implementing UserInterface.
 */
export class User implements UserInterface {
  public name: string;
  public email: string;

  constructor(name: string, email: string) {
    this.name = name;
    this.email = email;
  }

  /** Returns a greeting for the user. */
  public greet(greeting: string): string {
    return `${greeting}, ${this.name}!`;
  }

  private hashEmail(): string {
    return crypto.createHash('sha256').update(this.email).digest('hex');
  }
}

export class AdminUser extends User {
  role: Role;

  constructor(name: string, email: string) {
    super(name, email);
    this.role = Role.Admin;
  }
}

/**
 * Process an incoming request and return a response.
 */
export function handleRequest(req: Request, res: Response): void {
  res.json({ status: 'ok' });
}

/** Format a user name */
const formatName = (first: string, last: string): string => {
  return `${first} ${last}`;
};

function helperFunc(): void {
  // internal helper
}

export { formatName };

describe('User', () => {
  it('should greet correctly', () => {
    const user = new User('Alice', 'alice@test.com');
    expect(user.greet('Hello')).toBe('Hello, Alice!');
  });

  test('format name works', () => {
    expect(formatName('John', 'Doe')).toBe('John Doe');
  });
});
