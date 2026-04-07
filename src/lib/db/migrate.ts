import { migrate } from "drizzle-orm/libsql/migrator";

import { getDb } from "@/lib/db/client";

let migrationsRan = false;

export async function runMigrationsIfNeeded() {
  if (migrationsRan) return;

  const db = getDb();
  await migrate(db, { migrationsFolder: "./drizzle" });
  migrationsRan = true;
}
