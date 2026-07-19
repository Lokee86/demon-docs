# Space Rocks Mass-Rename Timing

Measured iterations: 5
All iterations valid: True

| Stage | Median | Mean | P95 | Min | Max |
|---|---:|---:|---:|---:|---:|
| copy_ms | 412.875 ms | 399.334 ms | 436.172 ms | 328.376 ms | 438.915 ms |
| baseline_initialize_ms | 3843.909 ms | 3987.783 ms | 4671.826 ms | 3684.105 ms | 4878.456 ms |
| baseline_repair_ms | 530.986 ms | 572.443 ms | 655.351 ms | 518.081 ms | 663.149 ms |
| baseline_check_ms | 170.199 ms | 174.872 ms | 201.759 ms | 157.816 ms | 208.763 ms |
| first_filesystem_rename_ms | 278.851 ms | 275.781 ms | 290.327 ms | 257.702 ms | 293.021 ms |
| first_precheck_ms | 1129.864 ms | 1168.528 ms | 1259.449 ms | 1112.708 ms | 1275.220 ms |
| first_fix_ms | 1928.290 ms | 1943.908 ms | 1992.826 ms | 1883.071 ms | 1992.919 ms |
| first_postcheck_ms | 1050.577 ms | 1064.981 ms | 1102.556 ms | 1042.105 ms | 1109.287 ms |
| first_idempotent_ms | 1191.803 ms | 1232.863 ms | 1330.399 ms | 1185.218 ms | 1350.349 ms |
| first_rename_cycle_ms | 5676.551 ms | 5686.060 ms | 5758.000 ms | 5583.293 ms | 5763.763 ms |
| second_filesystem_rename_ms | 247.729 ms | 248.946 ms | 256.627 ms | 240.088 ms | 257.673 ms |
| second_precheck_ms | 1176.859 ms | 1175.509 ms | 1214.554 ms | 1117.493 ms | 1220.245 ms |
| second_fix_ms | 1979.776 ms | 1986.692 ms | 2012.758 ms | 1967.426 ms | 2015.903 ms |
| second_postcheck_ms | 1051.612 ms | 1049.355 ms | 1087.939 ms | 1017.633 ms | 1094.588 ms |
| second_idempotent_ms | 1193.489 ms | 1225.436 ms | 1318.789 ms | 1140.133 ms | 1328.647 ms |
| second_rename_cycle_ms | 5722.121 ms | 5685.937 ms | 5792.135 ms | 5535.832 ms | 5798.626 ms |
| two_rename_cycles_ms | 11305.414 ms | 11371.997 ms | 11550.135 ms | 11212.382 ms | 11562.389 ms |
| total_ms | 16754.679 ms | 16855.069 ms | 17321.241 ms | 16624.808 ms | 17459.758 ms |

## Throughput

- First fix: 176.32 files/s; 1927.61 link repairs/s
- Second fix: 171.74 files/s; 1877.49 link repairs/s
- Complete two-cycle scenario: 60.33 renamed files/s; 657.56 applied link repairs/s
