import { request, type Page, type SpareReturnable, type StatusChange } from "./api";

export type ReportLayout = "summary" | "detailed";

export type ReportStep = {
  status: string;
  enteredAt: string;
  leftAt: string;
};

/** Timeline for the detailed report: oldest first. Leave time is empty for the current state. */
export function reportSteps(createdAt: string, status: string, history: StatusChange[]): ReportStep[] {
  if (history.length === 0) {
    return [{ status, enteredAt: createdAt, leftAt: "" }];
  }
  const chronological = [...history].reverse();
  const steps: ReportStep[] = [
    {
      status: chronological[0].from,
      enteredAt: createdAt,
      leftAt: chronological[0].createdAt,
    },
  ];
  for (let i = 0; i < chronological.length; i++) {
    const next = chronological[i + 1];
    steps.push({
      status: chronological[i].to,
      enteredAt: chronological[i].createdAt,
      leftAt: next ? next.createdAt : "",
    });
  }
  return steps;
}

/** Exit date only when the service is currently delivered. */
export function reportExitAt(status: string, history: StatusChange[]): string {
  if (status !== "delivered") {
    return "";
  }
  for (const change of history) {
    if (change.to === "delivered") {
      return change.createdAt;
    }
  }
  return "";
}

/** Every net spare part still on the service (uses minus returns). */
export async function loadReportSpares(serviceId: string, signal?: AbortSignal): Promise<SpareReturnable[]> {
  const items: SpareReturnable[] = [];
  let after: SpareReturnable | null = null;
  for (;;) {
    const params = new URLSearchParams({ limit: "100" });
    if (after) {
      params.set("after", after.brand);
      params.set("afterModel", after.model);
      params.set("afterId", after.id);
    }
    const page = await request<Page<SpareReturnable>>(`/api/services/${serviceId}/spare-returns?${params}`, { signal });
    items.push(...page.items);
    if (!page.hasMore || page.items.length === 0) {
      break;
    }
    after = page.items[page.items.length - 1];
  }
  return items;
}
