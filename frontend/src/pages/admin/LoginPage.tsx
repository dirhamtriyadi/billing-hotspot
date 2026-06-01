import { Navigate, useNavigate } from "react-router-dom";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { toast } from "sonner";
import { Loader2, Wifi } from "lucide-react";
import { useAuth } from "@/context/auth";
import { applyApiErrors, errorMessage } from "@/lib/form";
import { loginSchema, type LoginValues } from "@/schemas";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

export default function LoginPage() {
  const { user, login } = useAuth();
  const navigate = useNavigate();

  const form = useForm<LoginValues>({
    resolver: zodResolver(loginSchema),
    defaultValues: { username: "", password: "" },
  });

  if (user) return <Navigate to="/admin/dashboard" replace />;

  const onSubmit = async (values: LoginValues) => {
    try {
      await login(values.username, values.password);
      toast.success("Berhasil masuk");
      navigate("/admin/dashboard");
    } catch (e) {
      if (!applyApiErrors(e, form.setError)) toast.error(errorMessage(e));
    }
  };

  return (
    <div className="hero-gradient flex min-h-screen items-center justify-center p-4">
      <Card className="w-full max-w-sm">
        <CardHeader className="items-center text-center">
          <div className="mb-2 flex h-12 w-12 items-center justify-center rounded-xl bg-primary text-primary-foreground">
            <Wifi className="h-6 w-6" />
          </div>
          <CardTitle className="text-2xl">Admin Hotspot</CardTitle>
          <p className="text-sm text-muted-foreground">
            Masuk untuk mengelola billing
          </p>
        </CardHeader>
        <CardContent>
          <form
            onSubmit={form.handleSubmit(onSubmit)}
            className="space-y-4"
            noValidate
          >
            <div className="space-y-2">
              <Label htmlFor="username">Username</Label>
              <Input
                id="username"
                autoComplete="username"
                {...form.register("username")}
              />
              {form.formState.errors.username && (
                <p className="text-sm text-destructive">
                  {form.formState.errors.username.message}
                </p>
              )}
            </div>
            <div className="space-y-2">
              <Label htmlFor="password">Password</Label>
              <Input
                id="password"
                type="password"
                autoComplete="current-password"
                {...form.register("password")}
              />
              {form.formState.errors.password && (
                <p className="text-sm text-destructive">
                  {form.formState.errors.password.message}
                </p>
              )}
            </div>
            <Button
              type="submit"
              className="w-full"
              disabled={form.formState.isSubmitting}
            >
              {form.formState.isSubmitting && (
                <Loader2 className="h-4 w-4 animate-spin" />
              )}
              Masuk
            </Button>
            <p className="text-center text-xs text-muted-foreground">
              Default: <code>admin</code> / <code>admin123</code>
            </p>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
