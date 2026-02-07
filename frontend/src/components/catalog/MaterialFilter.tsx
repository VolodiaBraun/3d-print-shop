"use client";

import { Checkbox } from "@/components/ui/checkbox";

const MATERIALS = ["PLA", "PETG", "ABS", "TPU", "Resin", "Nylon"];

interface MaterialFilterProps {
  selectedMaterials: string[];
  onMaterialsChange: (materials: string[]) => void;
}

export function MaterialFilter({
  selectedMaterials,
  onMaterialsChange,
}: MaterialFilterProps) {
  const toggle = (material: string) => {
    if (selectedMaterials.includes(material)) {
      onMaterialsChange(selectedMaterials.filter((m) => m !== material));
    } else {
      onMaterialsChange([...selectedMaterials, material]);
    }
  };

  return (
    <div>
      <h3 className="text-sm font-semibold mb-3">Материал</h3>
      <div className="space-y-2">
        {MATERIALS.map((material) => (
          <label
            key={material}
            className="flex items-center gap-2 cursor-pointer text-sm text-muted-foreground hover:text-foreground transition-colors"
          >
            <Checkbox
              checked={selectedMaterials.includes(material)}
              onCheckedChange={() => toggle(material)}
            />
            {material}
          </label>
        ))}
      </div>
    </div>
  );
}
