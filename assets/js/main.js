
$(function () {
    /*=======================
                UI Slider Range JS
    =========================*/
    $("#slider-range").slider({
        range: true,
        min: 0,
        max: 2500,
        values: [10, 2500],
        slide: function (event, ui) {
            $("#amount").val("$" + ui.values[0] + " - $" + ui.values[1]);
        }
    });
    $("#amount").val("$" + $("#slider-range").slider("values", 0) +
        " - $" + $("#slider-range").slider("values", 1));

    let domShippingCalculationMsg = $("#shipping-calculation-msg");

    $(".province_id").change(function () {
        provinceID = $(".province_id").val()

        $(".city_id").find("option")
            .remove()
            .end()
            .append('<option value="">Pilih Kota / Kabupaten</option>')

        $.ajax({
            url: "/carts/cities?province_id=" + provinceID,
            method: "GET",
            success: function (result) {
                $.each(result.data, function (i, city) {
                    $(".city_id").append(`<option value="${city.city_id}">${city.city_name}</option>`)
                });
            }
        })
    });

    $(".city_id").change(function () {
        let cityID = $(".city_id").val()
        let courier = $(".courier").val()

        $(".shipping_fee_options").find("option")
            .remove()
            .end()
            .append('<option value="">Pilih Paket</option>')

        $.ajax({
            url: "/carts/calculate-shipping",
            method: "POST",
            data: {
                city_id: cityID,
                courier: courier
            },
            success: function (result) {
                domShippingCalculationMsg.html('');
                $.each(result.data, function (i, shipping_fee_option) {
                    $(".shipping_fee_options").append(`<option value="${shipping_fee_option.service}">${shipping_fee_option.fee} (${shipping_fee_option.service})</option>`);
                });
            },
            error: function (e) {
                domShippingCalculationMsg.html(`<div class="alert alert-warning">Perhitungan ongkos kirim gagal!</div>`);
            }
        })
    });

    $(".shipping_fee_options").change(function () {
        let cityID = $(".city_id").val()
        let courier = $(".courier").val()
        let shippingFee = $(this).val();

        $.ajax({
            url: "/carts/apply-shipping",
            method: "POST",
            data: {
                shipping_package: shippingFee.split("-")[0],
                city_id: cityID,
                courier: courier
            },
            success: function (result) {
                $("#grand-total").text(result.data.grand_total)
            },
            error: function (e) {
                domShippingCalculationMsg.html(`<div class="alert alert-warning">Pemilihan paket ongkir gagal!</div>`);
            }
        })
    });
});
// Global variable to store selected files
let selectedFiles = new Set();

// Function to handle file selection
function handleFiles(files) {
    const maxFileSize = 5 * 1024 * 1024; // 5MB
    const imagePreviewContainer = document.getElementById('imagePreviewContainer');

    Array.from(files).forEach(file => {
        // Validate file size
        if (file.size > maxFileSize) {
            alert(`File ${file.name} is too large. Maximum size is 5MB`);
            return;
        }

        // Validate file type
        if (!file.type.startsWith('image/')) {
            alert(`File ${file.name} is not an image`);
            return;
        }

        // Add to selected files
        selectedFiles.add(file);

        // Create preview
        const reader = new FileReader();
        reader.onload = function(e) {
            const previewDiv = document.createElement('div');
            previewDiv.className = 'image-preview';
            previewDiv.innerHTML = `
                        <img src="${e.target.result}" alt="${file.name}">
                        <span class="remove-image" data-filename="${file.name}">
                            <i class="fa-solid fa-xmark"></i>
                        </span>
                    `;
            imagePreviewContainer.appendChild(previewDiv);
        };
        reader.readAsDataURL(file);
    });

    // Update file input label
    updateFileInputLabel();
}

// Function to update file input label
function updateFileInputLabel() {
    const fileCount = selectedFiles.size;
    $('.custom-file-label').html(
        fileCount ? `${fileCount} file${fileCount > 1 ? 's' : ''} selected` : 'Choose files or drop them here'
    );
}

// File input change handler
$('#productImages').on('change', function(e) {
    handleFiles(this.files);
});

// Remove image handler
$(document).on('click', '.remove-image', function() {
    const filename = $(this).data('filename');
    selectedFiles = new Set([...selectedFiles].filter(file => file.name !== filename));
    $(this).closest('.image-preview').remove();
    updateFileInputLabel();
});

// Drag and drop handlers
const dropZone = document.getElementById('dropZone');

['dragenter', 'dragover', 'dragleave', 'drop'].forEach(eventName => {
    dropZone.addEventListener(eventName, preventDefaults, false);
});

function preventDefaults(e) {
    e.preventDefault();
    e.stopPropagation();
}

['dragenter', 'dragover'].forEach(eventName => {
    dropZone.addEventListener(eventName, highlight, false);
});

['dragleave', 'drop'].forEach(eventName => {
    dropZone.addEventListener(eventName, unhighlight, false);
});

function highlight(e) {
    dropZone.classList.add('dragover');
}

function unhighlight(e) {
    dropZone.classList.remove('dragover');
}

dropZone.addEventListener('drop', handleDrop, false);

function handleDrop(e) {
    const dt = e.dataTransfer;
    const files = dt.files;
    handleFiles(files);
}

// Form submission handler
$('#productForm').on('submit', function(e) {
    e.preventDefault();

    const formData = new FormData(this);

    // Remove existing files from FormData
    formData.delete('productImages[]');

    // Add selected files
    selectedFiles.forEach(file => {
        formData.append('productImages[]', file);
    });

    // Here you would normally send the formData to your server
    // For demonstration, let's log the files
    console.log('Files to upload:', Array.from(selectedFiles));
});