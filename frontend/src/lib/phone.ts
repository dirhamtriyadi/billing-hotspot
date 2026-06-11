// WhatsApp / Indonesian phone-number helpers, shared by the admin settings
// (validate + normalise at input) and the storefront (build wa.me links), so
// both use one consistent standard.

/**
 * Normalise an Indonesian number to the wa.me standard: country code first,
 * digits only, no "+" and no leading 0 (e.g. "6281313102678"). Accepts loose
 * input — "0812…", "+62 812…", "62-812…", "812…" — and returns "" when empty.
 */
export function normalizeWaNumber(raw: string | undefined): string {
  const digits = (raw || "").replace(/\D/g, ""); // strip +, spaces, dashes …
  if (!digits) return "";
  if (digits.startsWith("0")) return "62" + digits.slice(1);
  if (digits.startsWith("62")) return digits;
  if (digits.startsWith("8")) return "62" + digits; // missing country code
  return digits;
}

/**
 * Whether the (normalised) value looks like a real Indonesian WhatsApp number:
 * 62, then a mobile prefix 8, then 7–12 more digits (total 10–15 digits).
 */
export function isValidWaNumber(raw: string | undefined): boolean {
  return /^628\d{7,12}$/.test(normalizeWaNumber(raw));
}
