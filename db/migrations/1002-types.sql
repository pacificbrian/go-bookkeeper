-- +migrate Up

INSERT INTO `account_types` ('id', 'name') VALUES
  (1,'Cash'),(2,'Checking/Deposit'),(3,'Credit Card'),(4,'Investment'),
  (5,'Health Care'),(6,'Loan'),(7,'Asset'),(8,'Crypto'),(9,NULL);

INSERT INTO `cash_flow_types` ('id', 'name') VALUES
  (1,'Debit'),(2,'Credit'),(3,'Debit (Transfer)'),(4,'Credit (Transfer)');

INSERT INTO `category_types` ('id', 'name') VALUES
  (1,'Expense'),(2,'Income'),(3,'Additional Expense'),
  (4,'Uncommon Expense'), (5,'User Expense');

INSERT INTO `categories` VALUES
  (1,'--',0,0,NULL),(2,'Auto:Fuel',1,0,NULL),(3,'Auto:Parking',3,0,NULL),
  (4,'Auto:Service',1,0,NULL),(5,'Auto:Registration',3,0,NULL),
  (6,'Auto:Parts',4,0,NULL),(7,'Business',3,0,NULL),(8,'Cash',1,0,NULL),
  (9,'Charity',1,0,NULL),(10,'Unused',9,0,NULL),(11,'Clothing',1,0,NULL),
  (12,'Food:Dining',1,0,NULL),(13,'Education',1,0,NULL),
  (14,'Electronics',1,0,NULL),(15,'Entertainment',1,0,NULL),(16,'Unused',9,0,NULL),
  (17,'Fees:Bank',1,0,NULL),(18,'Fees:Other',1,0,NULL),(19,'Unused',9,0,NULL),
  (20,'Health:Fitness',3,0,NULL),(21,'Gifts',1,0,NULL),(22,'Food:Groceries',1,0,NULL),
  (23,'Unused',9,0,NULL),(24,'Home:Furnishings',1,0,NULL),
  (25,'Home:Furniture',4,0,NULL),(26,'Home:Improvement',1,0,NULL),
  (27,'Household',1,0,NULL),(28,'Insurance:Automotive',1,0,NULL),
  (29,'Insurance:Disability',3,0,NULL),(30,'Insurance:Life',1,0,NULL),
  (31,'Insurance:Medical',1,0,NULL),(32,'Insurance:Liability',4,0,NULL),
  (33,'Insurance:Property',1,0,NULL),(34,'Interest',1,0,NULL),
  (35,'Interest:Mortgage',1,1,NULL),(36,'Unused',9,0,NULL),
  (37,'Travel:Lodging',3,0,NULL),(38,'Medical:General',1,0,NULL),
  (39,'Medical:Dental',1,0,NULL),(40,'Unused',9,0,NULL),
  (41,'Medical:Pharmacy',1,0,NULL),(42,'Medical:Vision',3,0,NULL),
  (43,'Miscellaneous',1,0,NULL),(44,'Unused',9,0,NULL),
  (45,'Unused',9,0,NULL),(46,'Pet Care',1,0,NULL),(47,'Recreation',1,0,NULL),
  (48,'Rent',1,0,NULL),(49,'Shipping',3,0,NULL),(50,'Subscriptions',3,0,NULL),
  (51,'Taxes:Federal',1,1,NULL),(52,'Taxes:State',1,1,NULL),
  (53,'Taxes:Medicare',1,1,NULL),(54,'Taxes:Soc Sec',1,1,NULL),
  (55,'Taxes:SDI',1,1,NULL),(56,'Taxes:Property',1,0,NULL),
  (57,'Taxes:Foreign',3,0,NULL),(58,'Unused',9,0,NULL),
  (59,'Transportation',1,0,NULL),(60,'Travel',1,0,NULL),
  (61,'Utilities:Cable TV',9,0,NULL),(62,'Unused',9,0,NULL),
  (63,'Utilities:Energy',1,0,NULL),(64,'Unused',9,0,NULL),
  (65,'Utilities:Internet',9,0,NULL),(66,'Utilities:Telecomm',1,0,NULL),
  (67,'Utilities:Trash',1,0,NULL),(68,'Utilities:Water',1,0,NULL),
  (69,'Wages:Salary',2,0,NULL),(70,'Wages:Bonus',2,0,NULL),
  (71,'Business Income',2,0,NULL),(72,'Dividend',2,0,NULL),
  (73,'Gift Income',2,0,NULL),(74,'Interest Income',2,0,NULL),
  (75,'Investment Income',2,0,NULL),(76,'Other Income',2,0,NULL),
  (77,'Rent Income',2,0,NULL),(78,'Resale Income',2,0,NULL),
  (80,'Home:Appliance',3,0,0),(81,'Medical:Orthopedic',3,0,0),
  (83,'Child:Daycare',1,0,0),(84,'Health:Wellness',1,0,0),
  (85,'Home:Security',3,0,0),(86,'Home:Outdoor',3,0,0),
  (87,'Reimbursed Expenses',1,0,1),(88,'Child:Baby',3,0,0),
  (89,'Child:Toys&Etc',1,0,0),(90,'Health:Beauty',3,0,0),
  (91,'Food:Bakery&Cafe',3,0,0),(92,'Food:Alcohol',3,0,0),
  (93,'Auto:Lease',3,0,0),(94,'Wages:Restricted Stock',2,1,0),
  (95,'Support',1,1,0),(96,'Wages:Retirement',2,1,0),
  (97,'Taxes:Advance Credits',2,0,0);

INSERT INTO `currency_types` ('id', 'name', 'description') VALUES
  (1,'USD',NULL),(2,'AUD',NULL),(3,'BRL',NULL),(4,'CAD',NULL),(5,'CHF',NULL),
  (6,'EUR',NULL),(7,'GBP',NULL),(8,'JPY',NULL),(9,'NOK',NULL),(10,'NZD',NULL),
  (11,'SEK',NULL),(12,'XAU',NULL),(13,'XAG',NULL),(14,'ZAR',NULL);

INSERT INTO `repeat_interval_types` VALUES
  (1,'Once',0),(2,'Weekly',7),(3,'Bi-Weekly',14),(4,'Semi-Monthly',15),
  (5,'Monthly',30),(6,'Bi-Monthly',60),(7,'Quarterly',90),(8,'Bi-Annually',180),
  (9,'Annually',360),(10,'Thirds',120);

INSERT INTO `security_basis_types` VALUES
   (1,'FIFO'),(2,'Average'),(3,'NoImport');

INSERT INTO `security_types` VALUES
   (1,'Stock'),(2,'Mutual Fund'),(3,'Bond'),(4,'Bond Fund'),
   (5,'Money Market'),(6,'Foreign Currency'),(7,'Foreign Stock'),
   (8,'Foreign Stock Fund'),(9,'Foreign Bond'),(10,'Foreign Bond Fund'),
   (11,'Short Stock'),(12,'Short Fund'),(13,'Commodity Stock'),
   (14,'Commodity Fund'),(15,'Commodities'),(16,'Precious Metal'),
   (17,'Real Estate'),(18,'Trusts'),(19,'Options');

INSERT INTO `trade_types` VALUES
  (1,'Buy'),(2,'Sell'),(3,'Dividend'),(4,'Distribution'),
  (5,'Dividend (Reinvest)'),(6,'Distribution (Reinvest)'),
  (7,'Shares In'),(8,'Shares Out'),(9,'Split');

INSERT INTO `users` ('id', 'login') VALUES (1,'primary');

INSERT INTO `tax_categories` VALUES
   (1,1,69,NULL),(2,1,70,NULL),(3,2,74,NULL),(4,93,51,NULL),
   (5,4,NULL,3),(6,4,NULL,5),(7,9,NULL,4),(8,9,NULL,6),
   (9,9,NULL,2),(10,85,57,NULL),(11,57,56,NULL),(12,61,35,NULL),
   (13,56,52,NULL),(14,1,31,NULL);

INSERT INTO `tax_constants` VALUES
   (1,NULL,1500,1500,850,300,122500,61250,122500,122500,2500,1250,2500,2500,
    0.0200,0.2800,0.2500,3000,1500,3000,3000,112500,75000,150000,112500,
    175000,87500,175000,175000,0.0750,0.0200,100,0.1000,0.8000,0.0300,0.0250,
    0.2500,0.2600,5,25,3000,100000,0.1000,0.1500,0.2500,0.2800,0.3300,0.3500,
    0.3960,0.1500);

INSERT INTO `tax_filing_status` VALUES
  (1,'Single','S'),(2,'Married Filing Jointly','MFJ'),
  (3,'Married Filing Separately','MFS'),(4,'Head of Household','HH');

INSERT INTO `tax_items` VALUES
  (1,'Wages','TaxIncomeItem',1,1),(2,'Interest Taxable','TaxIncomeItem',1,NULL),
  (3,'Interest Exempt','TaxIncomeItem',NULL,NULL),(4,'Ordinary Dividends','TaxIncomeItem',1,NULL),
  (5,'Qualified Dividends','TaxIncomeItem',NULL,NULL),(6,'Refunds','TaxIncomeItem',NULL,NULL),
  (7,'Alimony','TaxIncomeItem',NULL,NULL),(8,'Business','TaxIncomeItem',NULL,NULL),
  (9,'Capital Gain','TaxIncomeItem',1,NULL),(10,'Other Gain','TaxIncomeItem',NULL,NULL),
  (11,'IRA Distributions','TaxIncomeItem',NULL,NULL),
  (12,'IRA Distributions Taxable','TaxIncomeItem',NULL,NULL),
  (13,'Pension Annuities','TaxIncomeItem',NULL,NULL),
  (14,'Pension Annuities Taxable','TaxIncomeItem',NULL,NULL),
  (15,'Supplemental Schedule E','TaxIncomeItem',NULL,NULL),
  (16,'Farm','TaxIncomeItem',NULL,NULL),
  (17,'Unemployment','TaxIncomeItem',NULL,NULL),
  (18,'Social Security','TaxIncomeItem',NULL,NULL),
  (19,'Social Security Taxable','TaxIncomeItem',NULL,NULL),
  (20,'Other Income','TaxIncomeItem',NULL,NULL),
  (21,'Short Asset Sales','TaxIncomeCapitalGainItem',NULL,NULL),
  (22,'Short Asset Gain','TaxIncomeCapitalGainItem',NULL,NULL),
  (23,'Short Other','TaxIncomeCapitalGainItem',NULL,NULL),
  (24,'Short Carryover','TaxIncomeCapitalGainItem',NULL,NULL),
  (25,'Short K1','TaxIncomeCapitalGainItem',NULL,NULL),
  (26,'Long Asset Sales','TaxIncomeCapitalGainItem',NULL,NULL),
  (27,'Long Asset Gain','TaxIncomeCapitalGainItem',NULL,NULL),
  (28,'Long Other','TaxIncomeCapitalGainItem',NULL,NULL),
  (29,'Long Carryover','TaxIncomeCapitalGainItem',NULL,NULL),
  (30,'Long K1','TaxIncomeCapitalGainItem',NULL,NULL),
  (31,'Long Distributions','TaxIncomeCapitalGainItem',NULL,NULL),
  (32,'Collectibles','TaxIncomeCapitalGainItem',NULL,NULL),
  (33,'Unrecaptured','TaxIncomeCapitalGainItem',NULL,NULL),
  (34,'Short Carryforward','TaxIncomeCapitalGainItem',NULL,NULL),
  (35,'Long Carryforward','TaxIncomeCapitalGainItem',NULL,NULL),
  (36,'Archer MSA','TaxDeductionForAGIItem',NULL,NULL),
  (37,'Business 2106','TaxDeductionForAGIItem',NULL,NULL),
  (38,'HSA Deduction','TaxDeductionForAGIItem',NULL,NULL),
  (39,'Moving Expenses','TaxDeductionForAGIItem',NULL,NULL),
  (40,'Self-Employment Tax','TaxDeductionForAGIItem',NULL,NULL),
  (41,'SE Qualified Retirement','TaxDeductionForAGIItem',NULL,NULL),
  (42,'SE Health Insurance','TaxDeductionForAGIItem',NULL,NULL),
  (43,'Early Withdrawal Penalty','TaxDeductionForAGIItem',NULL,NULL),
  (44,'Alimony','TaxDeductionForAGIItem',NULL,NULL),
  (45,'IRA Deduction','TaxDeductionForAGIItem',NULL,NULL),
  (46,'Student Loan Interest','TaxDeductionForAGIItem',NULL,NULL),
  (47,'Educator Expenses','TaxDeductionForAGIItem',NULL,NULL),
  (48,'Tutition Fees','TaxDeductionForAGIItem',NULL,NULL),
  (49,'Lost Jury Pay','TaxDeductionForAGIItem',NULL,NULL),
  (50,'Domestic 8903','TaxDeductionForAGIItem',NULL,NULL),
  (51,'Standard Deduction','TaxDeductionFromAGIItem',NULL,NULL),
  (52,'Itemized Deductions','TaxDeductionFromAGIItem',NULL,NULL),
  (53,'Exemption Amount','TaxDeductionFromAGIItem',NULL,NULL),
  (54,'Medical Dental','TaxDeductionItemizedItem',NULL,NULL),
  (55,'Medical Dental Allowed','TaxDeductionItemizedItem',NULL,NULL),
  (56,'State Local Income Taxes','TaxDeductionItemizedItem',5,NULL),
  (57,'Real Estate Taxes','TaxDeductionItemizedItem',5,NULL),
  (58,'Personal Property Taxes','TaxDeductionItemizedItem',NULL,NULL),
  (59,'Other Taxes','TaxDeductionItemizedItem',NULL,NULL),
  (60,'Mortgage Interest Points','TaxDeductionItemizedItem',NULL,NULL),
  (61,'Mortgage Interest Other','TaxDeductionItemizedItem',5,NULL),
  (62,'Qualified Mortgage Insurance Premiums','TaxDeductionItemizedItem',NULL,NULL),
  (63,'Mortgage Points Other','TaxDeductionItemizedItem',NULL,NULL),
  (64,'Investment Interest','TaxDeductionItemizedItem',NULL,NULL),
  (65,'Gifts Cash Check','TaxDeductionItemizedItem',NULL,NULL),
  (66,'Gifts Other','TaxDeductionItemizedItem',NULL,NULL),
  (67,'Gifts Carryover','TaxDeductionItemizedItem',NULL,NULL),
  (68,'Casualty Theft Losses','TaxDeductionItemizedItem',NULL,NULL),
  (69,'Casualty Theft Losses Allowed','TaxDeductionItemizedItem',NULL,NULL),
  (70,'Unreimbursed Employee Expenses','TaxDeductionItemizedItem',NULL,NULL),
  (71,'Tax Preparation Fees','TaxDeductionItemizedItem',NULL,NULL),
  (72,'Investment Other Expenses','TaxDeductionItemizedItem',NULL,NULL),
  (73,'Total Job Misc Allowed','TaxDeductionItemizedItem',NULL,NULL),
  (74,'Other Misc Gambling','TaxDeductionItemizedItem',NULL,NULL),
  (75,'Other Misc Casualty Theft Losses','TaxDeductionItemizedItem',NULL,NULL),
  (76,'Other Misc All Other','TaxDeductionItemizedItem',NULL,NULL),
  (77,'Income Tax','TaxTaxItem',6,NULL),
  (78,'Other Tax','TaxTaxItem',NULL,NULL),
  (79,'AMT','TaxTaxItem',NULL,NULL),
  (80,'SE','TaxTaxItem',NULL,NULL),
  (81,'Unreported Tips','TaxTaxItem',NULL,NULL),
  (82,'Additional Retirement','TaxTaxItem',NULL,NULL),
  (83,'Advance Earned Income','TaxTaxItem',NULL,NULL),
  (84,'Household Employment','TaxTaxItem',NULL,NULL),
  (85,'Foreign Tax','TaxCreditItem',7,NULL),
  (86,'Dependent Care','TaxCreditItem',NULL,NULL),
  (87,'Elderly Care','TaxCreditItem',NULL,NULL),
  (88,'Education','TaxCreditItem',NULL,NULL),
  (89,'Retirement Contributions','TaxCreditItem',NULL,NULL),
  (90,'Residental Energy','TaxCreditItem',NULL,NULL),
  (91,'Child Tax','TaxCreditItem',NULL,NULL),
  (92,'Other','TaxCreditItem',NULL,NULL),
  (93,'Federal Tax Withheld','TaxPaymentItem',8,NULL),
  (94,'Tax Prepayments','TaxPaymentItem',NULL,NULL),
  (95,'Earned Income Credit','TaxPaymentItem',NULL,NULL),
  (96,'Combat Pay','TaxPaymentItem',NULL,NULL),
  (97,'Excess Social Security','TaxPaymentItem',NULL,NULL),
  (98,'Child Tax Credit','TaxPaymentItem',NULL,NULL),
  (99,'Filing Extension','TaxPaymentItem',NULL,NULL),
  (100,'Telephone Excise','TaxPaymentItem',NULL,NULL),
  (101,'Minimum Tax','TaxPaymentItem',NULL,NULL),
  (102,'Other Payments','TaxPaymentItem',NULL,NULL),
  (103,'Homebuyer Credit','TaxPaymentItem',NULL,NULL),
  (104,'Stimulus','TaxPaymentItem',NULL,NULL);

INSERT INTO `tax_regions` VALUES
  (1,'Federal'),(2,'State');

INSERT INTO `tax_types` VALUES
  (1,'Income'),(2,'Income_Capital_Gain'),(3,'Deductions for AGI'),
  (4,'Deductions from AGI'),(5,'Itemized Deductions'),(6,'Tax'),
  (7,'Tax Credits'),(8,'Tax Payments');

INSERT INTO `tax_years` VALUES
  (1,2007,NULL,5350,5350,10700,7850,1300,1050,1050,1300,3400,1133,1.5000,156400,117300,234600,195500,156400,78200,156400,156400,3.0000,0.1500,0.0500,31850,31850,63700,42650,33750,22500,45000,33750,7825,31850,77100,160850,349700,7825,31850,77100,160850,349700,15650,63700,128500,195850,349700,11200,42650,110100,178350,349700,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL),
  (2,2008,NULL,5450,5450,10900,8000,1350,1050,1050,1350,3500,2333,3.0000,159950,119975,239950,199950,159950,79975,159950,159950,1.5000,0.1500,0.0000,32350,32350,65100,43650,46200,34975,69950,46200,8025,32550,78850,164550,357700,8025,32550,78850,164550,357700,16050,65100,131450,200300,357700,11450,43650,112650,182400,357700,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL),
  (3,2009,NULL,5700,5700,11400,8350,NULL,NULL,NULL,NULL,3650,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,8350,33950,82250,171550,372950,8350,33950,68525,104425,186475,16700,67900,137050,208850,372950,11950,45500,117450,190200,372950,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL),
  (4,2010,NULL,5700,5700,11400,8400,NULL,NULL,NULL,NULL,3650,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,8375,34000,82400,171850,373650,8375,34000,68650,104625,186825,16750,68000,137300,209250,373650,11950,45550,117650,190550,373650,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL),
  (5,2011,NULL,5800,5800,11600,8500,NULL,NULL,NULL,NULL,3700,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,8500,34500,83600,174400,379150,8500,34500,69675,106150,189575,17000,69000,139350,212300,379150,12150,46250,119400,193350,379150,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL),
  (7,2012,NULL,5950,5950,11900,8700,NULL,NULL,NULL,NULL,3800,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,8700,35350,85650,178650,388350,8700,35350,71350,108725,194175,17400,70700,142700,217450,388350,12400,47350,122300,198050,388350,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL),
  (8,2013,NULL,6100,6100,12200,8950,NULL,NULL,NULL,NULL,3900,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,8925,36250,87850,183250,398350,8925,36250,73200,111525,199175,17850,72500,146400,223050,398350,12750,48600,125450,203150,398350,400000,225000,450000,425000,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL),
  (9,2014,NULL,6200,6200,12400,9100,NULL,NULL,NULL,NULL,3950,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,9075,36900,89350,186350,405100,9075,36900,74425,113425,202550,18150,73800,148850,226850,405100,12950,49400,127550,206600,405100,406750,228800,457600,432200,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL),
  (10,2015,NULL,6300,6300,12600,9250,NULL,NULL,NULL,NULL,4000,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,9225,37450,90750,189300,411500,9225,37450,75600,115225,205750,18450,74900,151200,230450,411550,13150,50200,129600,209850,411500,413200,232425,464850,439000,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL),
  (11,2016,NULL,6300,6300,12600,9300,NULL,NULL,NULL,NULL,4050,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,9275,37650,91150,190150,413350,9275,37650,75950,115725,206675,18550,75300,151900,231450,413350,13250,50400,130150,210800,413350,415050,233475,466950,441100,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL),
  (12,2017,NULL,6350,6350,12700,9350,NULL,NULL,NULL,NULL,4050,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,9325,37950,91900,191650,416700,9325,37950,76550,116675,208350,18650,75900,153100,233350,416700,13350,50800,131200,212500,416700,418400,235350,470700,444550,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL),
  (13,2018,NULL,12000,12000,24000,18000,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,9525,38700,82500,157500,200000,9525,38700,82500,157500,200000,19050,77400,165000,315000,400000,13600,51800,82500,157500,200000,500000,300000,600000,500000,0.1000,0.1200,0.2200,0.2400,0.3200,0.3500,0.3700,10000),
  (14,2019,NULL,12200,12200,24400,18350,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,9700,39475,84200,160725,204100,9700,39475,84200,160725,204100,19400,78950,168400,321450,408200,13850,52850,84200,160700,204100,510300,306175,612350,510300,0.1000,0.1200,0.2200,0.2400,0.3200,0.3500,0.3700,10000),
  (15,2020,NULL,12400,12400,24800,18650,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,9875,40125,85525,163300,207350,9875,40125,85525,163300,207350,19750,80250,171050,326600,414700,14100,53700,85500,163300,207350,518400,311025,622050,518400,0.1000,0.1200,0.2200,0.2400,0.3200,0.3500,0.3700,10000),
  (16,2021,NULL,12550,12550,25100,18800,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,9950,40525,86375,164925,209425,9950,40525,86375,164925,209425,19900,81050,172750,329850,418850,14200,54200,86350,164900,209400,523600,314150,628300,523600,0.1000,0.1200,0.2200,0.2400,0.3200,0.3500,0.3700,10000);


-- +migrate Down
DELETE FROM `account_types`;
DELETE FROM `cash_flow_types`;
DELETE FROM `catgory_types`;
DELETE FROM `categories`;
DELETE FROM `currency_types`;
DELETE FROM `repeat_interval_types`;
DELETE FROM `security_basis_types`;
DELETE FROM `security_types`;
DELETE FROM `trade_types`;
DELETE FROM `users`;

DELETE FROM `tax_categories`;
DELETE FROM `tax_constants`;
DELETE FROM `tax_filing_status`;
DELETE FROM `tax_items`;
DELETE FROM `tax_regions`;
DELETE FROM `tax_types`;
DELETE FROM `tax_years`;
