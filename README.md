# fynance

*A calculator for auditing Sell-To-Cover transactions for Option Exercises.*

## Future Plans

- [ ] Incorporate an avenue for system data to be fed directly to the calculator, corresponding to a `stdout` stream of structured results for repurposing.
- [ ] Enable the full **CLI** interfacing alongside the Fyne UI, providing **dual functionality**
- [ ] Flesh-out all possible **inputs** to ensure full coverage of the process
- [ ] Add a version inclusive of **Releases** and **Restricted Stock** as well.

### Build

Run the following command to build *(hides that pesky terminal window)*...

```bash
go build -ldflags="-H=windowsgui" -o fynance.exe
```

---

## Calculating Releases Too

Sell To Covers can take place on more than just options. They can also occur during a release of shares. Here the rules are just a tad bit different. It helps to understand the actual commission and fee structure that is applied on a per/client basis...

*These are the following charges that may be levied against a specific client.* **No client is subject to ALL**:

- Commission (BPS), defaults to 0.03 on total transaction value.
- Commission (Cents), calculated per share, charges $XX.XX per share, usually a few cents.
- Minimum Commission, measured in dollars, usually $25 or $50.
- Wire Fee, usually $25/transaction
- Exercise Fee, per transaction

For Commission, clients technically can have BOTH versions applied to their transactions.
