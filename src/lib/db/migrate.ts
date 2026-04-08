import { migrate } from "drizzle-orm/libsql/migrator";
import { access } from "node:fs/promises";
import path from "node:path";

import { getDb } from "@/lib/db/client";

let migrationsRan = false;

export async function runMigrationsIfNeeded() {
  if (migrationsRan) return;

  const migrationsFolder = path.join(process.cwd(), "drizzle");

  try {
    await access(migrationsFolder);
  } catch {
    // In some serverless bundles the drizzle folder is not present at runtime.
    // Skip runtime migrations and let the app use the existing database schema.
    migrationsRan = true;
    console.warn(
      "[db] Skipping migrations: ./drizzle folder not found at runtime.",
    );
    return;
  }

  const db = getDb();
  await migrate(db, { migrationsFolder });
  migrationsRan = true;
}
