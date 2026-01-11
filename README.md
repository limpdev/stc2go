# stc2go

*A calculator for auditing Sell-To-Cover transactions for Option Exercises.*

#### Future Plans

- [] Incorporate an avenue for system data to be fed directly to the calculator, corresponding to a `stdout` stream of structured results for repurposing.
- [] Enable the full **CLI** interfacing alongside the Fyne UI, providing **dual functionality**
- [] Flesh-out all possible **inputs** to ensure full coverage of the process
- [] Add a version inclusive of **Releases** and **Restricted Stock** as well.

### Build

Run the following command to build *(hides that pesky terminal window)*...

```bash
go build -ldflags="-H=windowsgui" -o stc2go.exe
```