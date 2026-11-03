import { useState } from "react";
import { catalogImageSrc } from "./api";

export function ModelIcon({ brand, model }: { brand: string; model: string }) {
  const [failed, setFailed] = useState(false);
  if (failed) {
    return null;
  }
  return <img className="model-icon" alt="" src={catalogImageSrc(brand, model)} onError={() => setFailed(true)} />;
}
