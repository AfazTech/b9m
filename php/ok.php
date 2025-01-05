<?php

require 'api.php'; // مسیر فایل کلاس را وارد کنید

$dnsManager = new DNSManager('http://172.21.41.164:8080'); // URL API خود را وارد کنید

try {
    // افزودن دامنه
    //  $dnsManager->addDomain('afaz.me', 'ns58.servercap.com', 'ns59.servercap.com');
    // echo "Domain added successfully.\n";

    // افزودن رکورد
    // var_dump($dnsManager->addRecord('afaz.me', 'www', 'A', '192.0.2.1', 3600));
    // echo "Record added successfully.\n";

    // دریافت تمام رکوردها
    $records = $dnsManager->getAllRecords('afaz.me');
    print_r($records);

    // // حذف رکورد
    // $dnsManager->deleteRecord('example.com', 'www');
    // echo "Record deleted successfully.\n";

    // // حذف دامنه
    // $dnsManager->deleteDomain('example.com');
    // echo "Domain deleted successfully.\n";

} catch (Exception $e) {
    echo $e->getMessage();
}
