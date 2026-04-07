import { createInsertSchema, createSelectSchema } from "drizzle-zod";
import { z } from "zod";

import { logs, sessions } from "@/lib/db/schema";

export const insertSessionSchema = createInsertSchema(sessions, {
  locale: z.enum(["en", "fi"]),
  lastLevel: z.number().int().min(0).max(100),
});

export const selectSessionSchema = createSelectSchema(sessions);

export const insertLogSchema = createInsertSchema(logs, {
  locale: z.enum(["en", "fi"]),
  levelReached: z.number().int().min(0).max(100),
});

export const selectLogSchema = createSelectSchema(logs);

export type InsertSessionDto = z.infer<typeof insertSessionSchema>;
export type SelectSessionDto = z.infer<typeof selectSessionSchema>;
export type InsertLogDto = z.infer<typeof insertLogSchema>;
export type SelectLogDto = z.infer<typeof selectLogSchema>;
