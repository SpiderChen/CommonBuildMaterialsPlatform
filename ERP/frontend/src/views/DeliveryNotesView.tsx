import type { BootstrapData } from "../services/types";
import { ERPWorkbenchView } from "./ERPWorkbenchView";
import type { WorkbenchMenuItem } from "./ERPWorkbenchView";

type DeliveryNotesViewProps = {
  bootstrap: BootstrapData | null;
  menuItems: WorkbenchMenuItem[];
  selectedSiteId: number;
  onChanged: () => void;
};

export function DeliveryNotesView({ bootstrap, menuItems, selectedSiteId, onChanged }: DeliveryNotesViewProps) {
  return (
    <ERPWorkbenchView
      section="delivery"
      bootstrap={bootstrap}
      menuItems={menuItems}
      selectedSiteId={selectedSiteId}
      onChanged={onChanged}
    />
  );
}
