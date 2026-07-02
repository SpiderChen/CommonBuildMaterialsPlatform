import { CheckCircle2, ImagePlus } from "lucide-react";
import { type ChangeEvent, FormEvent, useEffect, useState } from "react";
import { Button, Field, LayoutRegion, LoginForm, Panel, StatusChip, TextAreaInput, TextInput, useMessageBox } from "../components";
import { api } from "../services/api";
import type { PublicDeliverySignDetail } from "../services/types";
import { browserFilePayload } from "../utils/filePayload";
import { sensitiveActionPrompt } from "../utils/sensitiveActions";

export function PublicSignView({ token }: { token: string }) {
  const [detail, setDetail] = useState<PublicDeliverySignDetail | null>(null);
  const [signer, setSigner] = useState("");
  const [phone, setPhone] = useState("");
  const [signedQty, setSignedQty] = useState("");
  const [photoURL, setPhotoURL] = useState("");
  const [photoChecksum, setPhotoChecksum] = useState("");
  const [fileName, setFileName] = useState("");
  const [remark, setRemark] = useState("");
  const [error, setError] = useState("");
  const [signedNo, setSignedNo] = useState("");
  const [geoPosition, setGeoPosition] = useState<{ longitude: number; latitude: number } | null>(null);
  const { showError, confirmMessage } = useMessageBox();

  async function load() {
    const next = await api.publicSignDetail(token);
    setDetail(next);
    setPhone(next.link.phone || next.order.phone || "");
    setSignedQty(String(next.dispatch.planQuantity || next.order.planQuantity || ""));
  }

  useEffect(() => {
    load().catch((err: unknown) => setError(err instanceof Error ? err.message : "签收链接不可用"));
  }, [token]);

  useEffect(() => {
    if (error) {
      showError(error, "签收失败");
    }
  }, [error, showError]);

  useEffect(() => {
    if (!navigator.geolocation) return;
    navigator.geolocation.getCurrentPosition(
      (position) => setGeoPosition({
        longitude: position.coords.longitude,
        latitude: position.coords.latitude
      }),
      () => setGeoPosition(null),
      { enableHighAccuracy: true, timeout: 8000, maximumAge: 60000 }
    );
  }, []);

  async function handlePhotoFileChange(event: ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0];
    if (!file) return;
    try {
      const payload = await browserFilePayload(file);
      setPhotoURL(payload.url);
      setPhotoChecksum(payload.checksum);
      setFileName(payload.fileName);
    } catch (err) {
      setError(err instanceof Error ? err.message : "读取现场照片失败");
    }
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const prompt = sensitiveActionPrompt("public-sign-submit", "提交配送签收");
    if (prompt) {
      const confirmed = await confirmMessage({
        title: prompt.title,
        message: prompt.message,
        tone: "warning",
        confirmLabel: prompt.confirmLabel,
        confirmVariant: prompt.confirmVariant
      });
      if (!confirmed) return;
    }
    setError("");
    try {
      const signed = await api.publicSign(token, {
        signer,
        phone,
        signedQty: Number(signedQty || detail?.dispatch.planQuantity || 0),
        ...(geoPosition ? { longitude: geoPosition.longitude, latitude: geoPosition.latitude } : {}),
        photo: photoURL,
        signature: signer ? `${signer} 电子签名` : "电子签名",
        remark,
        attachments: photoURL ? [{ fileName: fileName || "delivery-photo", fileType: "photo", url: photoURL, checksum: photoChecksum }] : []
      });
      setSignedNo(signed.signNo);
    } catch (err) {
      setError(err instanceof Error ? err.message : "签收失败");
    }
  }

  if (signedNo) {
    return (
      <LayoutRegion as="main" className="login-shell">
        <Panel className="login-card">
          <CheckCircle2 size={42} />
          <h1>签收完成</h1>
          <p className="muted">{signedNo}</p>
        </Panel>
      </LayoutRegion>
    );
  }

  return (
    <LayoutRegion as="main" className="login-shell">
      <Panel className="login-card">
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
        <LoginForm onSubmit={submit}>
          <Field label="签收人">
            <TextInput value={signer} onChange={(event) => setSigner(event.target.value)} required />
          </Field>
          <Field label="手机号">
            <TextInput value={phone} onChange={(event) => setPhone(event.target.value)} />
          </Field>
          <Field label="实收数量">
            <TextInput value={signedQty} onChange={(event) => setSignedQty(event.target.value)} required />
          </Field>
          <Field label="现场照片">
            <input type="file" accept="image/*" onChange={handlePhotoFileChange} />
          </Field>
          <Field label="附件名">
            <TextInput value={fileName} onChange={(event) => setFileName(event.target.value)} />
          </Field>
          <Field label="备注">
            <TextAreaInput value={remark} onChange={(event) => setRemark(event.target.value)} />
          </Field>
          <Button variant="primary" type="submit" icon={<ImagePlus size={16} />}>提交签收</Button>
        </LoginForm>
      </Panel>
    </LayoutRegion>
  );
}
