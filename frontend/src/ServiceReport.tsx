import { logoSrc, type Brand, type Device, type ServiceState, type SpareReturnable } from "./api";
import { formatWhen, statusLabel } from "./format";
import { shopName, type Copy, type Lang } from "./i18n";
import { Icon } from "./icons";
import { serviceQrSvg } from "./qr";
import { reportExitAt, reportSteps, type ReportLayout } from "./report";

type Props = {
  copy: Copy;
  lang: Lang;
  brand: Brand;
  states: ServiceState[];
  device: Device;
  spares: SpareReturnable[];
  layout: ReportLayout;
};

export function ServiceReport({ copy, lang, brand, states, device, spares, layout }: Props) {
  const exitAt = reportExitAt(device.status, device.history);
  const steps = layout === "detailed" ? reportSteps(device.createdAt, device.status, device.history) : [];
  const qr = serviceQrSvg(device.id);

  return (
    <article className="service-report" aria-hidden="true">
      <header className="service-report-head">
        <div className="service-report-brand">
          {brand.logo ? <img className="service-report-logo" src={logoSrc} alt="" /> : (
            <span className="service-report-mark" aria-hidden="true">
              <Icon name="services" size={18} />
            </span>
          )}
          <div>
            <strong>{shopName(copy, brand.name)}</strong>
            <p>
              {copy.serviceReport}
              {" · "}
              {layout === "summary" ? copy.reportSummary : copy.reportDetailed}
            </p>
          </div>
        </div>
        <div className="service-report-qr">
          <div className="service-report-qr-svg" dangerouslySetInnerHTML={{ __html: qr }} />
          <span className="mono">{device.id}</span>
        </div>
      </header>

      <section className="service-report-block">
        <h2>{copy.customer}</h2>
        <dl className="service-report-dl">
          <div>
            <dt>{copy.customerName}</dt>
            <dd>{device.customer.name}</dd>
          </div>
          <div>
            <dt>{copy.phone}</dt>
            <dd>{device.customer.phone}</dd>
          </div>
          {device.customer.email ? (
            <div>
              <dt>{copy.reportEmail}</dt>
              <dd>{device.customer.email}</dd>
            </div>
          ) : null}
          {device.customer.address ? (
            <div>
              <dt>{copy.reportAddress}</dt>
              <dd>{device.customer.address}</dd>
            </div>
          ) : null}
        </dl>
      </section>

      <section className="service-report-block">
        <h2>{copy.device}</h2>
        <dl className="service-report-dl">
          <div>
            <dt>{copy.brand}</dt>
            <dd>{device.brand}</dd>
          </div>
          <div>
            <dt>{copy.model}</dt>
            <dd>{device.model}</dd>
          </div>
          <div>
            <dt>{copy.serial}</dt>
            <dd className="mono">{device.serial}</dd>
          </div>
        </dl>
      </section>

      <section className="service-report-block">
        <h2>{copy.status}</h2>
        <dl className="service-report-dl">
          <div>
            <dt>{copy.registeredOn}</dt>
            <dd>{formatWhen(device.createdAt, lang)}</dd>
          </div>
          <div>
            <dt>{copy.reportExitDate}</dt>
            <dd>{exitAt ? formatWhen(exitAt, lang) : <span className="service-report-blank" />}</dd>
          </div>
          <div>
            <dt>{copy.reportCurrentStatus}</dt>
            <dd>{statusLabel(device.status, states, lang)}</dd>
          </div>
        </dl>
      </section>

      <section className="service-report-block">
        <h2>{copy.spareParts}</h2>
        {spares.length === 0 ? (
          <p className="service-report-empty">{copy.reportNoSpares}</p>
        ) : (
          <table className="service-report-table">
            <thead>
              <tr>
                <th>{copy.brand}</th>
                <th>{copy.model}</th>
                <th>{copy.quantity}</th>
                <th>{copy.reportAmount}</th>
              </tr>
            </thead>
            <tbody>
              {spares.map((part) => (
                <tr key={part.id}>
                  <td>{part.brand}</td>
                  <td>{part.model}</td>
                  <td>{part.returnable}</td>
                  <td>
                    <span className="service-report-blank" />
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </section>

      <section className="service-report-block">
        <table className="service-report-table service-report-fees">
          <tbody>
            <tr>
              <th scope="row">{copy.reportServiceFee}</th>
              <td>
                <span className="service-report-blank" />
              </td>
            </tr>
            <tr>
              <th scope="row">{copy.reportTotal}</th>
              <td>
                <span className="service-report-blank" />
              </td>
            </tr>
          </tbody>
        </table>
      </section>

      {layout === "detailed" ? (
        <section className="service-report-block">
          <h2>{copy.statusHistory}</h2>
          <table className="service-report-table">
            <thead>
              <tr>
                <th>{copy.status}</th>
                <th>{copy.reportEntered}</th>
                <th>{copy.reportLeft}</th>
              </tr>
            </thead>
            <tbody>
              {steps.map((step, index) => (
                <tr key={`${step.status}-${step.enteredAt}-${index}`}>
                  <td>{statusLabel(step.status, states, lang)}</td>
                  <td>{formatWhen(step.enteredAt, lang)}</td>
                  <td>{step.leftAt ? formatWhen(step.leftAt, lang) : "—"}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </section>
      ) : null}

      <section className="service-report-sign">
        <div>
          <span className="service-report-blank wide" />
          <p>{copy.reportCustomerSign}</p>
        </div>
        <div>
          <span className="service-report-blank wide" />
          <p>{copy.reportAuthorizedSign}</p>
        </div>
      </section>
    </article>
  );
}
