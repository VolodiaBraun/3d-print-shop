"use client";

import { useState, useRef, useEffect } from "react";
import Link from "next/link";
import {
  ClipboardList,
  Calculator,
  CreditCard,
  Package,
  ArrowLeft,
  Loader2,
  CheckCircle2,
  Upload,
  X,
  FileText,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useAuth } from "@/lib/auth-context";
import { submitCustomOrder, uploadCustomOrderFile, getProfile } from "@/lib/api";

const STEPS = [
  {
    icon: ClipboardList,
    title: "Опишите задачу",
    desc: "Заполните форму: расскажите, что нужно напечатать, укажите требования к материалу, цвету, точности. При желании прикрепите 3D-файл.",
  },
  {
    icon: Calculator,
    title: "Получите расчёт",
    desc: "Наш менеджер изучит заявку и свяжется с вами для уточнения деталей и расчёта стоимости.",
  },
  {
    icon: CreditCard,
    title: "Оплатите заказ",
    desc: "После согласования цены оплатите заказ онлайн или при получении — как вам удобнее.",
  },
  {
    icon: Package,
    title: "Получите изделие",
    desc: "Готовое изделие можно забрать самовывозом или мы доставим его в любой город России.",
  },
];

const ALLOWED_EXTENSIONS = [
  ".stl", ".obj", ".3mf", ".step", ".stp", ".zip",
  ".jpg", ".jpeg", ".png", ".webp", ".heic", ".heif",
];
const MAX_FILES = 5;
const MAX_FILE_SIZE = 50 * 1024 * 1024; // 50 MB

export default function CustomOrderPage() {
  const { isAuthenticated } = useAuth();

  const [name, setName] = useState("");
  const [phone, setPhone] = useState("");
  const [email, setEmail] = useState("");

  // Pre-fill from profile when authenticated
  useEffect(() => {
    if (!isAuthenticated) return;
    getProfile()
      .then((profile) => {
        if (profile.firstName || profile.lastName) {
          setName([profile.firstName, profile.lastName].filter(Boolean).join(" "));
        }
        if (profile.phone) setPhone(profile.phone);
        if (profile.email) setEmail(profile.email);
      })
      .catch(() => {});
  }, [isAuthenticated]);
  const [description, setDescription] = useState("");
  const [paymentMethod, setPaymentMethod] = useState("card");
  const [deliveryMethod, setDeliveryMethod] = useState("pickup");
  const [deliveryAddress, setDeliveryAddress] = useState("");
  const [files, setFiles] = useState<File[]>([]);
  const [fileError, setFileError] = useState("");
  const fileInputRef = useRef<HTMLInputElement>(null);

  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
  const [submitting, setSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState("");
  const [result, setResult] = useState<{ orderNumber: string } | null>(null);

  function addFiles(newFiles: FileList | null) {
    if (!newFiles) return;
    setFileError("");
    const toAdd: File[] = [];
    for (const file of Array.from(newFiles)) {
      const ext = "." + (file.name.split(".").pop()?.toLowerCase() ?? "");
      if (!ALLOWED_EXTENSIONS.includes(ext)) {
        setFileError(
          `Формат ${ext} не поддерживается. Допустимые: STL, OBJ, 3MF, STEP, ZIP`
        );
        return;
      }
      if (file.size > MAX_FILE_SIZE) {
        setFileError(`Файл «${file.name}» превышает 50 МБ`);
        return;
      }
      toAdd.push(file);
    }
    setFiles((prev) => {
      const combined = [...prev, ...toAdd];
      if (combined.length > MAX_FILES) {
        setFileError(`Максимум ${MAX_FILES} файлов`);
        return prev;
      }
      return combined;
    });
  }

  function removeFile(index: number) {
    setFiles((prev) => prev.filter((_, i) => i !== index));
    setFileError("");
  }

  function validate(): boolean {
    const errors: Record<string, string> = {};
    if (!name.trim()) errors.name = "Введите имя";
    if (!phone.trim()) errors.phone = "Введите телефон";
    else if (phone.replace(/\D/g, "").length < 10)
      errors.phone = "Некорректный номер телефона";
    if (deliveryMethod !== "pickup" && !deliveryAddress.trim())
      errors.deliveryAddress = "Введите адрес доставки";
    setFieldErrors(errors);
    return Object.keys(errors).length === 0;
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!validate()) return;
    setSubmitting(true);
    setSubmitError("");
    try {
      const order = await submitCustomOrder({
        customerName: name.trim(),
        customerPhone: phone.trim(),
        customerEmail: email.trim() || undefined,
        clientDescription: description.trim() || undefined,
        paymentMethod,
        deliveryMethod,
        deliveryAddress:
          deliveryMethod !== "pickup" ? deliveryAddress.trim() : undefined,
      });
      for (const file of files) {
        try {
          await uploadCustomOrderFile(order.id, file);
        } catch {
          // Non-critical: continue even if file upload fails
        }
      }
      setResult({ orderNumber: order.orderNumber });
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { message?: string } } })
        ?.response?.data?.message;
      setSubmitError(msg || "Не удалось отправить заявку. Попробуйте позже.");
    } finally {
      setSubmitting(false);
    }
  }

  if (result) {
    return (
      <div className="mx-auto max-w-2xl px-4 py-16 text-center">
        <CheckCircle2 className="mx-auto h-16 w-16 text-green-500" />
        <h1 className="mt-4 text-2xl font-bold">Заявка отправлена!</h1>
        <p className="mt-2 text-lg text-muted-foreground">
          Номер заявки:{" "}
          <span className="font-mono font-bold text-foreground">
            {result.orderNumber}
          </span>
        </p>
        <p className="mt-4 text-muted-foreground">
          Мы свяжемся с вами для уточнения деталей и расчёта стоимости в
          течение рабочего дня.
        </p>
        <div className="mt-6 flex flex-col items-center gap-3 sm:flex-row sm:justify-center">
          <Button asChild>
            <Link href="/catalog">Перейти в каталог</Link>
          </Button>
          <Button variant="outline" asChild>
            <Link href="/">На главную</Link>
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-4xl px-4 py-8">
      <div className="mb-6">
        <Button variant="ghost" size="sm" asChild>
          <Link href="/">
            <ArrowLeft className="mr-1 h-4 w-4" />
            На главную
          </Link>
        </Button>
      </div>

      <h1 className="text-2xl font-bold lg:text-3xl">Индивидуальный заказ</h1>
      <p className="mt-2 text-muted-foreground">
        Напечатаем 3D-модель по вашему файлу или разработаем деталь под вашу
        задачу.
      </p>

      {/* Process steps */}
      <div className="mt-10 grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
        {STEPS.map((step, i) => {
          const Icon = step.icon;
          return (
            <div
              key={i}
              className="flex flex-col gap-3 rounded-xl border border-border bg-card p-5"
            >
              <div className="flex items-center gap-3">
                <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-primary/10 text-primary">
                  <Icon className="h-5 w-5" />
                </div>
                <span className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                  Шаг {i + 1}
                </span>
              </div>
              <h3 className="font-semibold">{step.title}</h3>
              <p className="text-sm leading-relaxed text-muted-foreground">
                {step.desc}
              </p>
            </div>
          );
        })}
      </div>

      {/* Order form */}
      <form onSubmit={handleSubmit} className="mt-12">
        <div className="grid gap-8 lg:grid-cols-[1fr_360px]">
          {/* Left column */}
          <div className="space-y-8">
            {/* Contacts */}
            <section>
              <h2 className="text-lg font-semibold">Контактные данные</h2>
              <div className="mt-4 space-y-4">
                <div>
                  <label className="mb-1 block text-sm font-medium">
                    Имя <span className="text-destructive">*</span>
                  </label>
                  <Input
                    placeholder="Ваше имя"
                    value={name}
                    onChange={(e) => {
                      setName(e.target.value);
                      if (fieldErrors.name)
                        setFieldErrors((p) => ({ ...p, name: "" }));
                    }}
                    aria-invalid={!!fieldErrors.name}
                  />
                  {fieldErrors.name && (
                    <p className="mt-1 text-sm text-destructive">
                      {fieldErrors.name}
                    </p>
                  )}
                </div>
                <div>
                  <label className="mb-1 block text-sm font-medium">
                    Телефон <span className="text-destructive">*</span>
                  </label>
                  <Input
                    type="tel"
                    placeholder="+7 (900) 123-45-67"
                    value={phone}
                    onChange={(e) => {
                      setPhone(e.target.value);
                      if (fieldErrors.phone)
                        setFieldErrors((p) => ({ ...p, phone: "" }));
                    }}
                    aria-invalid={!!fieldErrors.phone}
                  />
                  {fieldErrors.phone && (
                    <p className="mt-1 text-sm text-destructive">
                      {fieldErrors.phone}
                    </p>
                  )}
                </div>
                <div>
                  <label className="mb-1 block text-sm font-medium">
                    Email
                  </label>
                  <Input
                    type="email"
                    placeholder="email@example.com"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                  />
                </div>
              </div>
            </section>

            {/* Description */}
            <section>
              <h2 className="text-lg font-semibold">Описание задачи</h2>
              <div className="mt-4">
                <label className="mb-1 block text-sm font-medium">
                  Что нужно напечатать?
                </label>
                <textarea
                  className="min-h-[120px] w-full resize-y rounded-md border border-border bg-background px-3 py-2 text-sm outline-none focus:border-primary focus:ring-2 focus:ring-primary/20"
                  placeholder="Опишите деталь, укажите требования к материалу, цвету, точности…"
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                />
              </div>
            </section>

            {/* Files */}
            <section>
              <h2 className="text-lg font-semibold">Файлы (необязательно)</h2>
              <p className="mt-1 text-sm text-muted-foreground">
                3D-файлы: STL, OBJ, 3MF, STEP, ZIP. Фото образца: JPG, PNG,
                WEBP. Максимум {MAX_FILES} файлов по 50 МБ.
              </p>
              <div className="mt-4 space-y-3">
                {files.map((file, i) => (
                  <div
                    key={i}
                    className="flex items-center gap-3 rounded-lg border border-border bg-card px-4 py-3"
                  >
                    <FileText className="h-5 w-5 shrink-0 text-muted-foreground" />
                    <div className="min-w-0 flex-1">
                      <p className="truncate text-sm font-medium">{file.name}</p>
                      <p className="text-xs text-muted-foreground">
                        {(file.size / 1024 / 1024).toFixed(1)} МБ
                      </p>
                    </div>
                    <button
                      type="button"
                      onClick={() => removeFile(i)}
                      className="shrink-0 text-muted-foreground transition-colors hover:text-destructive"
                    >
                      <X className="h-4 w-4" />
                    </button>
                  </div>
                ))}

                {files.length < MAX_FILES && (
                  <>
                    <input
                      ref={fileInputRef}
                      type="file"
                      accept=".stl,.obj,.3mf,.step,.stp,.zip,.jpg,.jpeg,.png,.webp,.heic,.heif,image/*"
                      multiple
                      className="sr-only"
                      onChange={(e) => addFiles(e.target.files)}
                    />
                    <button
                      type="button"
                      onClick={() => fileInputRef.current?.click()}
                      className="flex w-full items-center justify-center gap-2 rounded-lg border border-dashed border-border bg-muted/30 px-4 py-6 text-sm text-muted-foreground transition-colors hover:border-muted-foreground/30 hover:bg-muted/50"
                    >
                      <Upload className="h-5 w-5" />
                      Выбрать файл
                    </button>
                  </>
                )}

                {fileError && (
                  <p className="text-sm text-destructive">{fileError}</p>
                )}
              </div>
            </section>
          </div>

          {/* Right column */}
          <div className="space-y-6 lg:sticky lg:top-24 lg:self-start">
            {/* Payment */}
            <div className="rounded-lg border border-border bg-card p-6">
              <h2 className="mb-4 text-base font-semibold">Способ оплаты</h2>
              <div className="space-y-3">
                {[
                  { value: "card", label: "Банковская карта" },
                  { value: "cash", label: "Наличные при получении" },
                ].map((opt) => (
                  <label
                    key={opt.value}
                    className={`flex cursor-pointer items-center gap-3 rounded-lg border px-4 py-3 transition-colors ${
                      paymentMethod === opt.value
                        ? "border-primary bg-primary/5"
                        : "border-border hover:border-muted-foreground/30"
                    }`}
                  >
                    <input
                      type="radio"
                      name="payment"
                      value={opt.value}
                      checked={paymentMethod === opt.value}
                      onChange={() => setPaymentMethod(opt.value)}
                      className="sr-only"
                    />
                    <div
                      className={`flex h-4 w-4 items-center justify-center rounded-full border-2 ${
                        paymentMethod === opt.value
                          ? "border-primary"
                          : "border-muted-foreground/40"
                      }`}
                    >
                      {paymentMethod === opt.value && (
                        <div className="h-2 w-2 rounded-full bg-primary" />
                      )}
                    </div>
                    <span className="text-sm font-medium">{opt.label}</span>
                  </label>
                ))}
              </div>
            </div>

            {/* Delivery */}
            <div className="rounded-lg border border-border bg-card p-6">
              <h2 className="mb-4 text-base font-semibold">Способ получения</h2>
              <div className="space-y-3">
                {[
                  { value: "pickup", label: "Самовывоз", hint: "Бесплатно" },
                  {
                    value: "courier",
                    label: "Курьерская доставка",
                    hint: "Рассчитывается отдельно",
                  },
                  {
                    value: "pickup_point",
                    label: "Пункт выдачи",
                    hint: "Рассчитывается отдельно",
                  },
                ].map((opt) => (
                  <label
                    key={opt.value}
                    className={`flex cursor-pointer items-start gap-3 rounded-lg border px-4 py-3 transition-colors ${
                      deliveryMethod === opt.value
                        ? "border-primary bg-primary/5"
                        : "border-border hover:border-muted-foreground/30"
                    }`}
                  >
                    <input
                      type="radio"
                      name="delivery"
                      value={opt.value}
                      checked={deliveryMethod === opt.value}
                      onChange={() => setDeliveryMethod(opt.value)}
                      className="sr-only"
                    />
                    <div
                      className={`mt-0.5 flex h-4 w-4 shrink-0 items-center justify-center rounded-full border-2 ${
                        deliveryMethod === opt.value
                          ? "border-primary"
                          : "border-muted-foreground/40"
                      }`}
                    >
                      {deliveryMethod === opt.value && (
                        <div className="h-2 w-2 rounded-full bg-primary" />
                      )}
                    </div>
                    <div>
                      <div className="text-sm font-medium">{opt.label}</div>
                      <div className="text-xs text-muted-foreground">
                        {opt.hint}
                      </div>
                    </div>
                  </label>
                ))}
              </div>

              {deliveryMethod !== "pickup" && (
                <div className="mt-4">
                  <label className="mb-1 block text-sm font-medium">
                    Адрес <span className="text-destructive">*</span>
                  </label>
                  <Input
                    placeholder="Город, улица, дом, квартира"
                    value={deliveryAddress}
                    onChange={(e) => {
                      setDeliveryAddress(e.target.value);
                      if (fieldErrors.deliveryAddress)
                        setFieldErrors((p) => ({
                          ...p,
                          deliveryAddress: "",
                        }));
                    }}
                    aria-invalid={!!fieldErrors.deliveryAddress}
                  />
                  {fieldErrors.deliveryAddress && (
                    <p className="mt-1 text-sm text-destructive">
                      {fieldErrors.deliveryAddress}
                    </p>
                  )}
                </div>
              )}
            </div>

            {submitError && (
              <p className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
                {submitError}
              </p>
            )}

            <Button
              type="submit"
              className="w-full"
              size="lg"
              disabled={submitting}
            >
              {submitting ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Отправляем заявку…
                </>
              ) : (
                "Отправить заявку"
              )}
            </Button>

            <p className="text-center text-xs text-muted-foreground">
              Стоимость будет рассчитана после изучения заявки. Менеджер
              свяжется с вами в течение рабочего дня.
            </p>
          </div>
        </div>
      </form>
    </div>
  );
}
