import { renderSVG } from "./vendor/uqr.js";

/** SVG QR whose payload is the plain service UUID. */
export function serviceQrSvg(uuid: string) {
  return renderSVG(uuid, { border: 1, ecc: "M" });
}
