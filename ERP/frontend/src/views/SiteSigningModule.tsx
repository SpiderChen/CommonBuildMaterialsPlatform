import { FileSignature, Link2, Plus } from "lucide-react";
import { FormEvent, useEffect, useMemo, useState } from "react";
import { ActionGroup, Button, DataTable, Dialog, DialogForm, Field, FormActions, HeroDateField, Panel, SelectInput, TextAreaInput, TextInput, ViewStack, buildDataTableRowContextMenu, useMessageBox } from "../components";
import { api } from "../services/api";
import type { DeliverySign, DeliverySignLink, DispatchOrder } from "../services/types";
import { sensitiveActionPrompt } from "../utils/sensitiveActions";

type SiteSigningProps = {
  selectedSiteId: number;
  onChanged: () => void;
};

type SignForm = {
  dispatchId: string;
  signer: string;
  phone: string;
  signedQty: string;
  photo: string;
  signature: string;
  remark: string;
};

type LinkForm = {
  dispatchId: string;
  channel: string;
  phone: string;
  expiresAt: string;
};

function toNumber(value: string, fallback = 0) {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : fallback;
}

function toISO(value: string) {
  if (!value) return "";
  const parsed = new Date(value);
  return Number.isNaN(parsed.valueOf()) ? value : parsed.toISOString();
}

export function SiteSigningModule({ selectedSiteId, onChanged }: SiteSigningProps) {
  const [dispatches, setDispatches] = useState<DispatchOrder[]>([]);
  const [signs, setSigns] = useState<DeliverySign[]>([]);
  const [signLinks, setSignLinks] = useState<DeliverySignLink[]>([]);
  const [loading, setLoading] = useState(false);
  const [busy, setBusy] = useState("");
  const [error, setError] = useState("");
  const [openSignDialog, setOpenSignDialog] = useState(false);
  const [openLinkDialog, setOpenLinkDialog] = useState(false);
  const { showError, confirmMessage } = useMessageBox();

  const [signForm, setSignForm] = useState<SignForm>({
    dispatchId: "0",
    signer: "",
    phone: "",
    signedQty: "",
    photo: "",
    signature: "",
    remark: ""
  });
  const [linkForm, setLinkForm] = useState<LinkForm>({
    dispatchId: "0",
    channel: "sms",
    phone: "",
    expiresAt: ""
  });

  const scopedDispatches = useMemo(
    () => (selectedSiteId ? dispatches.filter((item) => item.siteId === selectedSiteId) : dispatches),
    [dispatches, selectedSiteId]
  );
  const dispatchMap = useMemo(() => {
    const map = new Map<number, DispatchOrder>();
    dispatches.forEach((item) => map.set(item.id, item));
    return map;
  }, [dispatches]);

  const scopedSigns = useMemo(() => {
    if (!selectedSiteId) return signs;
    return signs.filter((item) => dispatchMap.get(item.dispatchId)?.siteId === selectedSiteId);
  }, [selectedSiteId, dispatchMap, signs]);

  const scopedSignLinks = useMemo(() => {
    if (!selectedSiteId) return signLinks;
    return signLinks.filter((item) => dispatchMap.get(item.dispatchId)?.siteId === selectedSiteId);
  }, [selectedSiteId, dispatchMap, signLinks]);

  const scopedDispatchIds = useMemo(() => new Set(scopedDispatches.map((item) => item.id)), [scopedDispatches]);
  const signedDispatchIds = useMemo(() => new Set(scopedSigns.map((item) => item.dispatchId)), [scopedSigns]);
  const pendingDispatches = useMemo(
    () => scopedDispatches.filter((item) => !signedDispatchIds.has(item.id) && !["signed", "completed", "cancelled", "void"].includes(item.status)),
    [scopedDispatches, signedDispatchIds]
  );

  async function load() {
    setError("");
    setLoading(true);
    try {
      const [dispatchesData, signsData, signLinksData] = await Promise.all([
        api.dispatchOrders(),
        api.signs(),
        api.signLinks()
      ]);
      setDispatches(dispatchesData);
      setSigns(signsData.sort((a, b) => b.id - a.id));
      setSignLinks(signLinksData.sort((a, b) => b.id - a.id));
    } catch (err) {
      setError(err instanceof Error ? err.message : "加载签收数据失败");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    load().catch(() => {});
  }, []);

  useEffect(() => {
    if (error) {
      showError(error, "操作失败");
    }
  }, [error, showError]);

  useEffect(() => {
    const candidate = scopedDispatches[0];
    if (!candidate) {
      if (signForm.dispatchId !== "0") {
        setSignForm((value) => ({ ...value, dispatchId: "0", signedQty: "" }));
      }
      if (linkForm.dispatchId !== "0") {
        setLinkForm((value) => ({ ...value, dispatchId: "0" }));
      }
      return;
    }

    const nextDispatchId = String(candidate.id);
    const nextSignedQty = String(candidate.loadedQty || candidate.planQuantity || "");
    if (signForm.dispatchId === "0" || !scopedDispatchIds.has(Number(signForm.dispatchId))) {
      setSignForm((value) => ({ ...value, dispatchId: nextDispatchId, signedQty: nextSignedQty }));
    } else if (!signForm.signedQty) {
      setSignForm((value) => ({ ...value, signedQty: nextSignedQty }));
    }
    if (linkForm.dispatchId === "0" || !scopedDispatchIds.has(Number(linkForm.dispatchId))) {
      setLinkForm((value) => ({ ...value, dispatchId: nextDispatchId }));
    }
  }, [scopedDispatches, scopedDispatchIds, signForm.dispatchId, linkForm.dispatchId]);

  function dispatchLabel(item: DispatchOrder) {
    return `${item.dispatchNo} / #${item.orderId} / ${item.productName}`;
  }

  function openSignForDispatch(item: DispatchOrder, source?: Partial<DeliverySign>) {
    setSignForm({
      ...signForm,
      dispatchId: String(item.id),
      signer: source?.signer || signForm.signer,
      phone: source?.phone || signForm.phone,
      signedQty: String(source?.signedQty || item.loadedQty || item.planQuantity || signForm.signedQty || "")
    });
    setOpenSignDialog(true);
  }

  function openLinkForDispatch(item: DispatchOrder, phone = "") {
    setLinkForm({
      ...linkForm,
      dispatchId: String(item.id),
      phone: phone || linkForm.phone
    });
    setOpenLinkDialog(true);
  }

  async function mutate(label: string, action: () => Promise<unknown>) {
    const prompt = sensitiveActionPrompt(label);
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
    setBusy(label);
    setError("");
    try {
      await action();
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "操作失败");
    } finally {
      setBusy("");
    }
  }

  async function submitSign(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const dispatchId = toNumber(signForm.dispatchId);
    if (!dispatchId) {
      setError("请先选择一条配送单");
      return;
    }
    await mutate("创建签收记录", () =>
      api.signDelivery({
        dispatchId,
        signer: signForm.signer,
        phone: signForm.phone,
        signedQty: toNumber(signForm.signedQty),
        photo: signForm.photo,
        signature: signForm.signature,
        remark: signForm.remark
      })
    );
    setOpenSignDialog(false);
  }

  async function submitLink(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const dispatchId = toNumber(linkForm.dispatchId);
    if (!dispatchId) {
      setError("请先选择一条配送单");
      return;
    }
    const payload: { dispatchId: number; channel: string; phone: string; expiresAt?: string } = {
      dispatchId,
      channel: linkForm.channel || "sms",
      phone: linkForm.phone
    };
    const expiresAt = toISO(linkForm.expiresAt);
    if (expiresAt) {
      payload.expiresAt = expiresAt;
    }
    await mutate("创建签收链接", () => api.createSignLink(payload));
    setOpenLinkDialog(false);
  }

  return (
    <ViewStack as="section">
      <Panel>
        <DataTable
          data={scopedSigns}
          rowKey={(item) => item.id}
          emptyText={loading ? "加载中..." : "暂无签收记录"}
          onRefresh={load}
          refreshDisabled={loading}
          rowContextMenu={buildDataTableRowContextMenu<DeliverySign>({
            actions: [
              {
                key: "create-sign-for-dispatch",
                label: "基于该配送单新增签收",
                disabled: () => busy !== "",
                onSelect: (item) => {
                  const dispatchItem = dispatchMap.get(item.dispatchId);
                  if (dispatchItem) openSignForDispatch(dispatchItem, item);
                }
              },
              {
                key: "create-link-for-dispatch",
                label: "为该配送单生成链接",
                disabled: () => busy !== "",
                onSelect: (item) => {
                  const dispatchItem = dispatchMap.get(item.dispatchId);
                  if (dispatchItem) openLinkForDispatch(dispatchItem, item.phone);
                }
              }
            ],
            copyFields: [
              { key: "sign", label: "签收单", value: (item) => item.signNo },
              { key: "dispatch", label: "配送单", value: (item) => {
                const dispatchItem = dispatchMap.get(item.dispatchId);
                return dispatchItem ? dispatchLabel(dispatchItem) : item.dispatchId;
              } },
              { key: "signer", label: "签收人", value: (item) => item.signer },
              { key: "phone", label: "手机号", value: (item) => item.phone }
            ]
          })}
          headerLeftAction={
            <ActionGroup>
              <Button variant="primary" icon={<FileSignature size={15} />} disabled={busy !== "" || !scopedDispatches.length} onClick={() => setOpenSignDialog(true)}>新增签收</Button>
              <Button variant="primary" icon={<Plus size={15} />} disabled={busy !== "" || !scopedDispatches.length} onClick={() => setOpenLinkDialog(true)}>生成签收链接</Button>
            </ActionGroup>
          }
          columns={[
            { key: "signNo", title: "签收单", render: (item) => item.signNo },
            {
              key: "dispatch",
              title: "配送单",
              render: (item) => {
                const dispatchItem = dispatchMap.get(item.dispatchId);
                return dispatchItem ? dispatchLabel(dispatchItem) : item.dispatchId;
              }
            },
            { key: "signer", title: "签收人", render: (item) => item.signer || "-" },
            { key: "signedQty", title: "签收方量", render: (item) => item.signedQty },
            { key: "signedAt", title: "签收时间", render: (item) => item.signedAt || "-" }
          ]}
        />
      </Panel>

      <Panel>
        <DataTable
          data={pendingDispatches}
          rowKey={(item) => item.id}
          title="待签收配送单"
          emptyText={loading ? "加载中..." : "暂无待签收配送单"}
          onRefresh={load}
          refreshDisabled={loading}
          rowContextMenu={buildDataTableRowContextMenu<DispatchOrder>({
            actions: [
              { key: "create-sign", label: "新增签收", disabled: () => busy !== "", onSelect: (item) => openSignForDispatch(item) },
              { key: "create-link", label: "生成签收链接", disabled: () => busy !== "", onSelect: (item) => openLinkForDispatch(item) }
            ],
            copyFields: [
              { key: "dispatch", label: "配送单", value: (item) => item.dispatchNo },
              { key: "order", label: "订单", value: (item) => item.orderId },
              { key: "product", label: "产品", value: (item) => item.productName }
            ]
          })}
          columns={[
            { key: "dispatchNo", title: "配送单", render: (item) => item.dispatchNo },
            { key: "orderId", title: "订单", render: (item) => `#${item.orderId}` },
            { key: "product", title: "产品", render: (item) => item.productName || "-" },
            { key: "qty", title: "计划/装载/签收", render: (item) => `${item.planQuantity} / ${item.loadedQty} / ${item.signedQty}` },
            { key: "eta", title: "预计到达", render: (item) => item.eta || "-" },
            { key: "status", title: "状态", render: (item) => item.status || "-" },
            {
              key: "actions",
              title: "操作",
              render: (item) => (
                <ActionGroup>
                  <Button size="sm" icon={<FileSignature size={13} />} disabled={busy !== ""} onClick={() => openSignForDispatch(item)}>签收</Button>
                  <Button size="sm" icon={<Link2 size={13} />} disabled={busy !== ""} onClick={() => openLinkForDispatch(item)}>链接</Button>
                </ActionGroup>
              )
            }
          ]}
        />
      </Panel>

      <Panel>
        <DataTable
          data={scopedSignLinks}
          rowKey={(item) => item.id}
          title="签收链接"
          emptyText={loading ? "加载中..." : "暂无签收链接"}
          onRefresh={load}
          refreshDisabled={loading}
          rowContextMenu={buildDataTableRowContextMenu<DeliverySignLink>({
            actions: [
              {
                key: "create-sign-for-link",
                label: "基于链接配送单新增签收",
                disabled: () => busy !== "",
                onSelect: (item) => {
                  const dispatchItem = dispatchMap.get(item.dispatchId);
                  if (dispatchItem) openSignForDispatch(dispatchItem, { phone: item.phone });
                }
              }
            ],
            copyFields: [
              { key: "link", label: "链接号", value: (item) => item.linkNo },
              { key: "url", label: "签收地址", value: (item) => item.url },
              { key: "token", label: "签收令牌", value: (item) => item.token },
              { key: "phone", label: "手机号", value: (item) => item.phone }
            ]
          })}
          columns={[
            { key: "linkNo", title: "链接号", render: (item) => item.linkNo },
            {
              key: "dispatch",
              title: "配送单",
              render: (item) => {
                const dispatchItem = dispatchMap.get(item.dispatchId);
                return dispatchItem ? dispatchLabel(dispatchItem) : item.dispatchId;
              }
            },
            { key: "channel", title: "渠道", render: (item) => `${item.channel || "-"} / ${item.phone || "-"}` },
            { key: "url", title: "地址", render: (item) => item.url || item.qrCode || "-" },
            { key: "expiresAt", title: "过期时间", render: (item) => item.expiresAt || "-" },
            { key: "status", title: "状态", render: (item) => item.status || "-" }
          ]}
        />
      </Panel>

      <Dialog open={openSignDialog} title="新增签收记录" className="master-dialog" closeDisabled={busy !== ""} onClose={() => setOpenSignDialog(false)}>
            <DialogForm
              onSubmit={async (event) => {
                await submitSign(event);
                if (!error) {
                  setOpenSignDialog(false);
                }
              }}
            >
              <Field label="配送单">
                <SelectInput
                  value={signForm.dispatchId}
                  onChange={(event) => setSignForm({ ...signForm, dispatchId: event.target.value })}
                  required
                >
                  <option value="0">请选择配送单</option>
                  {scopedDispatches.map((item) => (
                    <option key={item.id} value={item.id}>
                      {dispatchLabel(item)}
                    </option>
                  ))}
                </SelectInput>
              </Field>
              <Field label="签收人"><TextInput value={signForm.signer} onChange={(event) => setSignForm({ ...signForm, signer: event.target.value })} required /></Field>
              <Field label="手机号"><TextInput value={signForm.phone} onChange={(event) => setSignForm({ ...signForm, phone: event.target.value })} /></Field>
              <Field label="签收方量"><TextInput value={signForm.signedQty} onChange={(event) => setSignForm({ ...signForm, signedQty: event.target.value })} required /></Field>
              <Field label="现场照片"><TextInput value={signForm.photo} onChange={(event) => setSignForm({ ...signForm, photo: event.target.value })} /></Field>
              <Field label="电子签名"><TextInput value={signForm.signature} onChange={(event) => setSignForm({ ...signForm, signature: event.target.value })} /></Field>
              <Field label="备注" spanAll><TextAreaInput value={signForm.remark} onChange={(event) => setSignForm({ ...signForm, remark: event.target.value })} /></Field>
              <FormActions>
                <Button disabled={busy !== ""} onClick={() => setOpenSignDialog(false)}>取消</Button>
                <Button variant="primary" type="submit" icon={<FileSignature size={14} />} disabled={busy !== "" || !toNumber(signForm.dispatchId)}>保存签收</Button>
              </FormActions>
            </DialogForm>
      </Dialog>

      <Dialog open={openLinkDialog} title="新增签收链接" className="master-dialog" closeDisabled={busy !== ""} onClose={() => setOpenLinkDialog(false)}>
            <DialogForm
              onSubmit={async (event) => {
                await submitLink(event);
                if (!error) {
                  setOpenLinkDialog(false);
                }
              }}
            >
              <Field label="配送单">
                <SelectInput
                  value={linkForm.dispatchId}
                  onChange={(event) => setLinkForm({ ...linkForm, dispatchId: event.target.value })}
                  required
                >
                  <option value="0">请选择配送单</option>
                  {scopedDispatches.map((item) => (
                    <option key={item.id} value={item.id}>
                      {dispatchLabel(item)}
                    </option>
                  ))}
                </SelectInput>
              </Field>
              <Field label="发送渠道"><TextInput value={linkForm.channel} onChange={(event) => setLinkForm({ ...linkForm, channel: event.target.value })} /></Field>
              <Field label="手机号"><TextInput value={linkForm.phone} onChange={(event) => setLinkForm({ ...linkForm, phone: event.target.value })} /></Field>
              <HeroDateField label="过期时间" mode="date-time" value={linkForm.expiresAt} onChange={(expiresAt) => setLinkForm({ ...linkForm, expiresAt })} />
              <FormActions>
                <Button disabled={busy !== ""} onClick={() => setOpenLinkDialog(false)}>取消</Button>
                <Button variant="primary" type="submit" icon={<Plus size={14} />} disabled={busy !== "" || !toNumber(linkForm.dispatchId)}>生成链接</Button>
              </FormActions>
            </DialogForm>
      </Dialog>
    </ViewStack>
  );
}
