type Greeting = {
  message: string;
};

export function greeting(name: string): Greeting {
  return { message: `Hello, ${name}!` };
}
