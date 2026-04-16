export type ConsentCategory =
  | "essential"
  | "functional"
  | "analytics"
  | "marketing";

export type ConsentCategoryPreferences = Record<ConsentCategory, boolean>;

export type ConsentState = {
  version: number;
  updatedAt: string;
  categories: ConsentCategoryPreferences;
};
