/** Vendored from https://github.com/unjs/uqr (MIT). */
export function encode(
  data: string | ReadonlyArray<number>,
  options?: { ecc?: "L" | "M" | "Q" | "H"; border?: number },
): { version: number; size: number; data: boolean[][] };

export function renderSVG(
  data: string | ReadonlyArray<number>,
  options?: { ecc?: "L" | "M" | "Q" | "H"; border?: number },
): string;
