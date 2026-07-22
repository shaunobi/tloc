export interface User {
  id: string;
  displayName: string;
  roles: readonly string[];
}

type ApiError = {
  status: number;
  message: string;
};

export async function loadUser(baseUrl: URL, id: string): Promise<User> {
  const endpoint = new URL(`/v1/users/${encodeURIComponent(id)}`, baseUrl);
  const response = await fetch(endpoint, {
    headers: { accept: "application/json" },
  });
  if (!response.ok) {
    const error = (await response.json()) as ApiError;
    throw new Error(`${error.status}: ${error.message}`);
  }
  const value = (await response.json()) as User;
  return { ...value, roles: [...value.roles].sort() };
}

export const isAdministrator = (user: User): boolean =>
  user.roles.some((role) => role.toLowerCase() === "admin");
