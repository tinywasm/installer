# Windows Integration Testing — Setup Guide

VM: **win10** en KVM/libvirt (Debian 13)
Snapshot base: **windows-update**

---

## 1. Habilitar OpenSSH en Windows (una sola vez)

Conéctate a la VM por RDP desde Debian:

```bash
xfreerdp /v:192.168.122.150 /u:user /p:user /dynamic-resolution
```

> Si no tienes xfreerdp: `apt install freerdp2-x11`

Dentro de Windows, abre **PowerShell como Administrador** y ejecuta:

```powershell
# Instalar OpenSSH Server (incluido en Win10/11, solo hay que activarlo)
Add-WindowsCapability -Online -Name OpenSSH.Server~~~~0.0.1.0

# Iniciar y configurar inicio automático
Set-Service -Name sshd -StartupType Automatic
Start-Service sshd

# Usar PowerShell como shell por defecto (en lugar de cmd)
New-ItemProperty -Path "HKLM:\SOFTWARE\OpenSSH" `
  -Name DefaultShell `
  -Value "C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe" `
  -PropertyType String -Force

# Regla de firewall (normalmente se crea automático, por si acaso)
New-NetFirewallRule -Name "OpenSSH-Server" -DisplayName "OpenSSH Server" `
  -Protocol TCP -LocalPort 22 -Action Allow -Direction Inbound
```

---

## 2. Verificar desde Debian

```bash
# Puerto 22 abierto
(echo >/dev/tcp/192.168.122.150/22) 2>/dev/null && echo "OPEN" || echo "closed"

# Conexión SSH funcional
ssh user@192.168.122.150 "powershell.exe -Command 'Write-Host ok'"
# Esperado: ok
```

Si pide contraseña y quieres evitarla en los tests:

```bash
ssh-copy-id user@192.168.122.150
```

O instalar `sshpass` en host debian para pasar la contraseña vía `.env`:

```bash
apt install sshpass
```

---

## 3. Tomar nuevo snapshot con SSH habilitado

Una vez verificado que SSH funciona, tomar un snapshot limpio desde Debian:

```bash
# El snapshot debe tomarse con la VM corriendo
virsh -c qemu:///system snapshot-create-as win10 windows-update \
  --description "windows-update + OpenSSH habilitado" \
  --force
```

> `--force` sobreescribe si ya existe uno con ese nombre.

Verificar que quedó:

```bash
virsh -c qemu:///system snapshot-list win10
```

---

## 4. Configurar el test

```bash
cd /installer
cp .env.example .env
```

Editar `.env`:

```
WIN_VM_NAME=win10
WIN_SNAPSHOT_NAME=windows-update
WIN_IP=192.168.122.150
WIN_USER=user
WIN_PASS=user
```

---

## 5. Correr el test

```bash
cd ~/tinywasm/installer
go test -v -tags integration -run TestInstallScript ./...
```

Salida esperada:

```
=== RUN   TestWindowsConnectivity
    Step 1: revert to clean snapshot
    reverted win10 → windows-update
    Step 2: wait for Windows SSH
    SSH reachable at 192.168.122.150:22 ✅
    Step 3: create test file on desktop
    PS> New-Item -Path "C:\Users\user\Desktop\test_installer.txt" ...
    Step 4: verify file exists
    file exists ✅
    Step 5: revert to clean snapshot
    reverted win10 → windows-update
    Step 6: wait for Windows SSH
    SSH reachable at 192.168.122.150:22 ✅
    Step 7: verify file is gone after revert
    file is gone after revert ✅ — clean state confirmed
--- PASS: TestWindowsConnectivity
```

---

## Referencia rápida — comandos virsh útiles

```bash
# Ver todas las VMs
virsh -c qemu:///system list --all

# Ver snapshots de win10
virsh -c qemu:///system snapshot-list win10

# Revertir manualmente
virsh -c qemu:///system snapshot-revert win10 windows-update

# Ver IP asignada
ip neigh show | grep "52:54:00:9d:a4:22"
```
