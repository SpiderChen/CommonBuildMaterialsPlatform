import { CheckCircle2, ImagePlus } from "lucide-react";
import { FormEvent, useEffect, useState } from "react";
import { StatusChip } from "../components/StatusChip";
import { api } from "../services/api";
import type { PublicDeliverySignDetail } from "../services/types";

export function PublicSignView({ token }: { token: string }) {
  const [detail, setDetail] = useState<PublicDeliverySignDetail | null>(null);
  const [signer, setSigner] = useState("");
  const [phone, setPhone] = useState("");
  const [signedQty, setSignedQty] = useState("");
  const [photoURL, setPhotoURL] = useState("");
  const [fileName, setFileName] = useState("site-delivery-photo.jpg");
  const [remark, setRemark] = useState("现场验收无异议");
  const [error, setError] = useState("");
  const [signedNo, setSignedNo] = useState("");

  async function load() {
    const next = await api.publicSignDetail(token);
    setDetail(next);
    setPhone(next.link.phone || next.order.phone || "");
    setSignedQty(String(next.dispatch.planQuantity || next.order.planQuantity || ""));
  }

  useEffect(() => {
    load().catch((err: unknown) => setError(err instanceof Error ? err.message : "签收链接不可用"));
  }, [token]);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    try {
      const signed = await api.publicSign(token, {
        signer,
        phone,
        signedQty: Number(signedQty || detail?.dispatch.planQuantity || 0),
        longitude: 113.9452,
        latitude: 22.5358,
        photo: photoURL,
        signature: signer ? `${signer} 电子签名` : "电子签名",
        remark,
        attachments: photoURL ? [{ fileName, fileType: "photo", url: photoURL, checksum: "" }] : []
      });
      setSignedNo(signed.signNo);
    } catch (err) {
      setError(err instanceof Error ? err.message : "签收失败");
    }
  }

  if (signedNo) {
    return (
      <main className="login-shell">
        <section className="login-card panel">
          <CheckCircle2 size={42} />
          <h1>签收完成</h1>
          <p className="muted">{signedNo}</p>
        </section>
      </main>
    );
  }

  return (
    <main className="login-shell">
      <section className="login-card panel">
        <p className="eyebrow">工地签收</p>
        <h1>{detail?.dispatch.dispatchNo || "配送签收"}</h1>
        {detail ? (
          <div className="kpi-grid compact">
            <div><span>客户</span><b>{detail.customer}</b></div>
            <div><span>项目</span><b>{detail.project}</b></div>
            <div><span>产品</span><b>{detail.product}</b></div>
            <div><span>车牌</span><b>{detail.plateNo}</b></div>
            <div><span>数量</span><b>{detail.dispatch.planQuantity}</b></div>
            <div><span>状态</span><b><StatusChip value={detail.link.status} /></b></div>
          </div>
        ) : null}
        <form onSubmit={submit} className="login-form">
          <label>
            <span>签收人</span>
            <input value={signer} onChange={(event) => setSigner(event.target.value)} required />
          </label>
          <label>
            <span>手机号</span>
            <input value={phone} onChange={(event) => setPhone(event.target.value)} />
          </label>
          <label>
            <span>实收数量</span>
            <input value={signedQty} onChange={(event) => setSignedQty(event.target.value)} />
          </label>
          <label>
            <span>现场照片 URL</span>
            <input value={photoURL} onChange={(event) => setPhotoURL(event.target.value)} placeholder="minio://delivery/site-photo.jpg" />
          </label>
          <label>
            <span>附件名</span>
            <input value={fileName} onChange={(event) => setFileName(event.target.value)} />
          </label>
          <label>
            <span>备注</span>
            <textarea value={remark} onChange={(event) => setRemark(event.target.value)} />
          </label>
          <button className="primary-button" type="submit">
            <ImagePlus size={16} />
            提交签收
          </button>
          {error ? <p className="error-text">{error}</p> : null}
        </form>
      </section>
    </main>
  );
}
